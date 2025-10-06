//go:build never

package integration

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtual-staging-ai/api/internal/config"
	httpLib "github.com/virtual-staging-ai/api/internal/http"
	"github.com/virtual-staging-ai/api/internal/image"
	"github.com/virtual-staging-ai/api/internal/job"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
	workerEvents "github.com/virtual-staging-ai/worker/internal/events"
	workerQueue "github.com/virtual-staging-ai/worker/internal/queue"
	workerRepo "github.com/virtual-staging-ai/worker/internal/repository"
)

func setupTestDB(t *testing.T) *storage.DefaultDatabase {
	cfg, err := config.Load()
	require.NoError(t, err)

	db, err := storage.NewDefaultDatabase(&cfg.DB)
	require.NoError(t, err)
	return db
}

func TestEndToEnd_ImageLifecycle_SSE_Error(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("REDIS_ADDR not set; integration infra must start redis-test")
	}

	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	truncateAndSeed(t, db.Pool())

	ctx := context.Background()

	s3Svc, err := storage.NewDefaultS3Service(ctx, "vsa-it-bucket")
	require.NoError(t, err)
	require.NoError(t, s3Svc.CreateBucket(ctx))

	imgRepo := image.NewDefaultRepository(db)
	jobRepo := job.NewDefaultRepository(db)
	imgSvc := image.NewDefaultService(imgRepo, jobRepo)

	apiServer := httpLib.NewTestServer(db, s3Svc, imgSvc)
	ts := httptest.NewServer(apiServer)
	defer ts.Close()

	reqBody := createImageReq{
		ProjectID:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
		OriginalURL: "http://example.com/original2.jpg",
	}
	b, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/images", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)
	var created image.Image
	require.NoError(t, json.NewDecoder(res.Body).Decode(&created))
	_ = res.Body.Close()

	sseReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/events?image_id="+created.ID.String(), nil)
	sseRes, err := http.DefaultClient.Do(sseReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, sseRes.StatusCode)

	go func() {
		wctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		qc, err := workerQueue.NewAsynqQueueClient(&cfg)
		if err != nil {
			return
		}
		pub, _ := workerEvents.NewDefaultPublisherFromEnv()
		sqldb := openSQLDBFromEnv(t)
		defer sqldb.Close()
		imgWrite := workerRepo.NewImageRepository(sqldb)

		deadline := time.Now().Add(8 * time.Second)
		for time.Now().Before(deadline) {
			job, _ := qc.GetNextJob(wctx)
			if job == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			var payload struct {
				ImageID     string `json:"image_id"`
				OriginalURL string `json:"original_url"`
			}
			_ = json.Unmarshal(job.Payload, &payload)
			_ = imgWrite.SetProcessing(wctx, payload.ImageID)
			if pub != nil {
				_ = pub.PublishJobUpdate(wctx, workerEvents.JobUpdateEvent{JobID: job.ID, ImageID: payload.ImageID, Status: "processing"})
			}
			// Simulate failure
			errMsg := "processor failed"
			_ = imgWrite.SetError(wctx, payload.ImageID, errMsg)
			if pub != nil {
				_ = pub.PublishJobUpdate(wctx, workerEvents.JobUpdateEvent{JobID: job.ID, ImageID: payload.ImageID, Status: "error", Error: errMsg})
			}
			_ = qc.MarkJobFailed(wctx, job.ID, errMsg)
			return
		}
	}()

	r := bufio.NewReader(sseRes.Body)
	defer sseRes.Body.Close()
	var gotConnected, gotProcessing, gotError bool
	deadline := time.Now().Add(5 * time.Second)
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
				dataLine, _ := r.ReadString('\n')
				if strings.Contains(dataLine, `"status":"processing"`) {
					gotProcessing = true
				}
				if strings.Contains(dataLine, `"status":"error"`) {
					gotError = true
					break
				}
			}
		}
	}

	assert.True(t, gotConnected, "expected connected event")
	assert.True(t, gotProcessing, "expected processing status event")
	assert.True(t, gotError, "expected error status event")

	q := queries.New(db)
	parsedID, err := uuid.Parse(created.ID.String())
	require.NoError(t, err)
	dbImg, err := q.GetImageByID(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	require.NoError(t, err)
	assert.Equal(t, "error", string(dbImg.Status))
	// error text should be set
	assert.True(t, dbImg.Error.Valid)
}

func truncateAndSeed(t *testing.T, pool storage.PgxPool) {
	ctx := context.Background()
	// Truncate all tables (same as apps/api/tests/integration/storage_fixtures.go)
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE processed_events, invoices, subscriptions, images, jobs, projects, users, plans RESTART IDENTITY CASCADE
	`)
	require.NoError(t, err)
	// Seed minimal user+project (same values as seed.sql)
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, auth0_sub, stripe_customer_id, role) VALUES
		('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'auth0|testuser', 'cus_test', 'user');
	`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO projects (id, user_id, name) VALUES
		('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Test Project 1');
	`)
	require.NoError(t, err)
}

func openSQLDBFromEnv(t *testing.T) *sql.DB {
	host := getenv("PGHOST", "localhost")
	port := getenv("PGPORT", "5433")
	user := getenv("PGUSER", "testuser")
	pass := getenv("PGPASSWORD", "testpassword")
	dbname := getenv("PGDATABASE", "testdb")
	sslmode := getenv("PGSSLMODE", "disable")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, dbname, sslmode)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

type createImageReq struct {
	ProjectID   string  `json:"project_id"`
	OriginalURL string  `json:"original_url"`
	RoomType    *string `json:"room_type,omitempty"`
	Style       *string `json:"style,omitempty"`
	Seed        *int64  `json:"seed,omitempty"`
}

func TestEndToEnd_ImageLifecycle_SSE(t *testing.T) {
	// Ensure env is set for LocalStack S3 and Redis-backed queue
	t.Setenv("APP_ENV", "test")
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("REDIS_ADDR not set; integration infra must start redis-test")
	}

	// Setup DB and seed
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	truncateAndSeed(t, db.Pool())

	ctx := context.Background()

	// S3 service for API server
	s3Svc, err := storage.NewDefaultS3Service(ctx, "vsa-it-bucket")
	require.NoError(t, err)
	require.NoError(t, s3Svc.CreateBucket(ctx))

	// API services & server
	imgRepo := image.NewDefaultRepository(db)
	jobRepo := job.NewDefaultRepository(db)
	imgSvc := image.NewDefaultService(imgRepo, jobRepo)

	apiServer := httpLib.NewTestServer(db, s3Svc, imgSvc)
	ts := httptest.NewServer(apiServer)
	defer ts.Close()

	// Create image via API (this enqueues a stage:run task via asynq)
	reqBody := createImageReq{
		ProjectID:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
		OriginalURL: "http://example.com/original.jpg",
	}
	b, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/images", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)
	var created image.Image
	require.NoError(t, json.NewDecoder(res.Body).Decode(&created))
	_ = res.Body.Close()

	// Start SSE stream for this image
	sseReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/events?image_id="+created.ID.String(), nil)
	sseRes, err := http.DefaultClient.Do(sseReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, sseRes.StatusCode)

	// Start minimal worker loop to consume the enqueued task and publish SSE via Redis
	// (using worker queue + events + repository)
	go func() {
		wctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Queue client (asynq-backed)
		qc, err := workerQueue.NewAsynqQueueClient(&cfg)
		if err != nil {
			return
		}
		// Events publisher (with retry)
		pub, _ := workerEvents.NewDefaultPublisherFromEnv()
		// SQL DB for worker repository writes
		sqldb := openSQLDBFromEnv(t)
		defer sqldb.Close()
		imgWrite := workerRepo.NewImageRepository(sqldb)

		deadline := time.Now().Add(8 * time.Second)
		for time.Now().Before(deadline) {
			job, _ := qc.GetNextJob(wctx)
			if job == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			// Decode payload
			var payload struct {
				ImageID     string `json:"image_id"`
				OriginalURL string `json:"original_url"`
			}
			_ = json.Unmarshal(job.Payload, &payload)
			// Update DB: processing
			_ = imgWrite.SetProcessing(wctx, payload.ImageID)
			if pub != nil {
				_ = pub.PublishJobUpdate(wctx, workerEvents.JobUpdateEvent{JobID: job.ID, ImageID: payload.ImageID, Status: "processing"})
			}
			// Simulate work then ready
			time.Sleep(200 * time.Millisecond)
			staged := payload.OriginalURL + "-staged.jpg"
			_ = imgWrite.SetReady(wctx, payload.ImageID, staged)
			if pub != nil {
				_ = pub.PublishJobUpdate(wctx, workerEvents.JobUpdateEvent{JobID: job.ID, ImageID: payload.ImageID, Status: "ready"})
			}
			_ = qc.MarkJobCompleted(wctx, job.ID)
			return
		}
	}()

	// Read SSE lines and assert ordering
	r := bufio.NewReader(sseRes.Body)
	defer sseRes.Body.Close()
	var gotConnected, gotProcessing, gotReady bool
	deadline := time.Now().Add(5 * time.Second)
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
				// next data line contains status json
				dataLine, _ := r.ReadString('\n')
				if strings.Contains(dataLine, `"status":"processing"`) {
					gotProcessing = true
				}
				if strings.Contains(dataLine, `"status":"ready"`) {
					gotReady = true
					break
				}
			}
		}
	}

	assert.True(t, gotConnected, "expected connected event")
	assert.True(t, gotProcessing, "expected processing status event")
	assert.True(t, gotReady, "expected ready status event")

	// Verify DB updated to ready with staged_url
	q := queries.New(db)
	parsedID, err := uuid.Parse(created.ID.String())
	require.NoError(t, err)
	dbImg, err := q.GetImageByID(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	require.NoError(t, err)
	assert.Equal(t, "ready", string(dbImg.Status))
	if dbImg.StagedUrl.Valid {
		assert.True(t, strings.HasSuffix(dbImg.StagedUrl.String, "-staged.jpg"))
	}
}
