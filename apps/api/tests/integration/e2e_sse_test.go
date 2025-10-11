//go:build integration

package integration

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/api/internal/config"
	httpLib "github.com/real-staging-ai/api/internal/http"
	"github.com/real-staging-ai/api/internal/image"
	"github.com/real-staging-ai/api/internal/job"
	"github.com/real-staging-ai/api/internal/queue"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

func startAsynqWorker(t *testing.T, db storage.Database, emitError bool) (stop func()) {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	require.NotEmpty(t, addr, "REDIS_ADDR must be set for integration test")

	// Asynq server processing default queue
	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: addr}, asynq.Config{Concurrency: 1, Queues: map[string]int{"default": 1}})

	// Redis client for publishing SSE messages
	rdb := redis.NewClient(&redis.Options{Addr: addr})

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TaskTypeStageRun, func(ctx context.Context, task *asynq.Task) error {
		// Parse payload
		var payload struct {
			ImageID     string `json:"image_id"`
			OriginalURL string `json:"original_url"`
		}
		_ = json.Unmarshal(task.Payload(), &payload)

		q := queries.New(db)
		// processing
		parsed, _ := uuid.Parse(payload.ImageID)
		_, _ = q.UpdateImageStatus(ctx, queries.UpdateImageStatusParams{ID: pgtype.UUID{Bytes: parsed, Valid: true}, Status: queries.ImageStatus("processing")})
		_ = rdb.Publish(ctx, fmt.Sprintf("jobs:image:%s", payload.ImageID), `{"status":"processing"}`).Err()

		if emitError {
			// error path
			_, _ = q.UpdateImageWithError(ctx, queries.UpdateImageWithErrorParams{ID: pgtype.UUID{Bytes: parsed, Valid: true}, Error: text("processor failed")})
			_ = rdb.Publish(ctx, fmt.Sprintf("jobs:image:%s", payload.ImageID), `{"status":"error"}`).Err()
			return fmt.Errorf("fail to trigger retry")
		}

		// simulate work then ready
		time.Sleep(100 * time.Millisecond)
		staged := payload.OriginalURL + "-staged.jpg"
		_, _ = q.UpdateImageWithStagedURL(ctx, queries.UpdateImageWithStagedURLParams{ID: pgtype.UUID{Bytes: parsed, Valid: true}, StagedUrl: text(staged), Status: queries.ImageStatus("ready")})
		_ = rdb.Publish(ctx, fmt.Sprintf("jobs:image:%s", payload.ImageID), `{"status":"ready"}`).Err()
		return nil
	})

	go func() { _ = srv.Run(mux) }()
	return func() {
		srv.Shutdown()
		_ = rdb.Close()
	}
}

func text(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }

func newAPITestServer(t *testing.T, db *storage.DefaultDatabase) (*httptest.Server, storage.S3Service) {
	cfg, err := config.Load()
	require.NoError(t, err)
	ctx := context.Background()
	s3 := SetupTestS3Service(t, ctx)
	s3.Cfg.BucketName = "vsa-it-bucket"

	imgRepo := image.NewDefaultRepository(db)
	jobRepo := job.NewDefaultRepository(db)
	imgSvc := image.NewDefaultService(cfg, imgRepo, jobRepo)

	srv := httpLib.NewTestServer(db, s3, imgSvc)
	return httptest.NewServer(srv), s3
}

func TestE2E_SSE_ProcessingReady(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("REDIS_ADDR not set; integration infra must start redis-test")
	}
	// DB setup
	db := SetupTestDatabase(t)
	defer db.Close()
	require.NoError(t, ResetDatabase(context.Background(), db.Pool()))

	stop := startAsynqWorker(t, db, false)
	defer stop()

	ts, _ := newAPITestServer(t, db)
	defer ts.Close()

	// Create image
	reqBody := map[string]any{"project_id": "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", "original_url": "http://example.com/one.jpg"}
	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/images", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)
	var created map[string]any
	_ = json.NewDecoder(res.Body).Decode(&created)
	_ = res.Body.Close()
	imageID := created["id"].(string)

	// SSE stream
	sseReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/events?image_id="+imageID, nil)
	sseRes, err := http.DefaultClient.Do(sseReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, sseRes.StatusCode)
	defer sseRes.Body.Close()

	r := bufio.NewReader(sseRes.Body)
	deadline := time.Now().Add(5 * time.Second)
	var gotConnected, gotProcessing, gotReady bool
	for time.Now().Before(deadline) {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if strings.HasPrefix(line, "event: ") {
			if strings.Contains(line, "connected") {
				gotConnected = true
			}
			if strings.Contains(line, "job_update") {
				data, _ := r.ReadString('\n')
				if strings.Contains(data, `"status":"processing"`) {
					gotProcessing = true
				}
				if strings.Contains(data, `"status":"ready"`) {
					gotReady = true
					break
				}
			}
		}
	}
	assert.True(t, gotConnected)
	assert.True(t, gotProcessing)
	assert.True(t, gotReady)
}

func TestE2E_SSE_ProcessingError(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("REDIS_ADDR not set; integration infra must start redis-test")
	}
	// DB setup
	db := SetupTestDatabase(t)
	defer db.Close()
	require.NoError(t, ResetDatabase(context.Background(), db.Pool()))

	stop := startAsynqWorker(t, db, true)
	defer stop()

	ts, _ := newAPITestServer(t, db)
	defer ts.Close()

	// Create image
	reqBody := map[string]any{"project_id": "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", "original_url": "http://example.com/two.jpg"}
	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/images", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)
	var created map[string]any
	_ = json.NewDecoder(res.Body).Decode(&created)
	_ = res.Body.Close()
	imageID := created["id"].(string)

	// SSE stream
	sseReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/events?image_id="+imageID, nil)
	sseRes, err := http.DefaultClient.Do(sseReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, sseRes.StatusCode)
	defer sseRes.Body.Close()

	r := bufio.NewReader(sseRes.Body)
	deadline := time.Now().Add(5 * time.Second)
	var gotConnected, gotProcessing, gotError bool
	for time.Now().Before(deadline) {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if strings.HasPrefix(line, "event: ") {
			if strings.Contains(line, "connected") {
				gotConnected = true
			}
			if strings.Contains(line, "job_update") {
				data, _ := r.ReadString('\n')
				if strings.Contains(data, `"status":"processing"`) {
					gotProcessing = true
				}
				if strings.Contains(data, `"status":"error"`) {
					gotError = true
					break
				}
			}
		}
	}
	assert.True(t, gotConnected)
	assert.True(t, gotProcessing)
	assert.True(t, gotError)
}
