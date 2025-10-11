//go:build integration

package integration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/api/internal/storage/queries"
)

type presignReq struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
}

type presignResp struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresIn int64  `json:"expires_in"`
}

// RUN_E2E_UPLOAD_READY=1 to enable this end-to-end test.
func TestE2E_Presign_Upload_CreateImage_ReadyViaSSE(t *testing.T) {
	if os.Getenv("RUN_E2E_UPLOAD_READY") != "1" {
		t.Skip("set RUN_E2E_UPLOAD_READY=1 to run this test")
	}
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("REDIS_ADDR not set; integration infra must start redis-test")
	}
	// Test env for S3 service (uses LocalStack at http://localhost:4566)
	t.Setenv("APP_ENV", "test")

	ctx := context.Background()
	// DB setup
	db := SetupTestDatabase(t)
	defer db.Close()
	require.NoError(t, ResetDatabase(ctx, db.Pool()))

	// Start a minimal asynq worker that updates DB and publishes SSE events.
	stop := startAsynqWorker(t, db, false)
	defer stop()

	// Start API test server with real S3 service
	ts, _ := newAPITestServer(t, db)
	defer ts.Close()

	// 1) Presign
	preq := presignReq{Filename: "e2e.jpg", ContentType: "image/jpeg", FileSize: 1024}
	b, _ := json.Marshal(preq)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/uploads/presign", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	var p presignResp
	require.NoError(t, json.NewDecoder(res.Body).Decode(&p))
	_ = res.Body.Close()
	require.NotEmpty(t, p.UploadURL)
	require.NotEmpty(t, p.FileKey)

	// 2) Upload to presigned URL
	putReq, _ := http.NewRequest(http.MethodPut, p.UploadURL, bytes.NewReader([]byte("hello world")))
	putReq.Header.Set("Content-Type", preq.ContentType)
	putRes, err := http.DefaultClient.Do(putReq)
	require.NoError(t, err)
	// LocalStack may return 200 or 204
	assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, putRes.StatusCode)
	_ = putRes.Body.Close()

	// 3) Create image pointing to uploaded object (use a plausible public URL form)
	origURL := "https://test-bucket.s3.amazonaws.com/" + p.FileKey
	imgBody := map[string]any{"project_id": "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", "original_url": origURL}
	imgPayload, _ := json.Marshal(imgBody)
	imgReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/images", bytes.NewReader(imgPayload))
	imgReq.Header.Set("Content-Type", "application/json")
	imgRes, err := http.DefaultClient.Do(imgReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, imgRes.StatusCode)
	var created map[string]any
	require.NoError(t, json.NewDecoder(imgRes.Body).Decode(&created))
	_ = imgRes.Body.Close()
	imageID := created["id"].(string)

	// 4) SSE: wait for connected -> processing -> ready
	sseReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/events?image_id="+imageID, nil)
	sseRes, err := http.DefaultClient.Do(sseReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, sseRes.StatusCode)
	defer sseRes.Body.Close()

	r := bufio.NewReader(sseRes.Body)
	deadline := time.Now().Add(10 * time.Second)
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
	assert.True(t, gotConnected, "expected connected event")
	assert.True(t, gotProcessing, "expected processing event")
	assert.True(t, gotReady, "expected ready event")

	// 5) Verify DB row updated to ready with staged_url suffix
	q := queries.New(db)
	parsedID, err := uuid.Parse(imageID)
	require.NoError(t, err)
	dbImg, err := q.GetImageByID(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	require.NoError(t, err)
	assert.Equal(t, "ready", string(dbImg.Status))
	if dbImg.StagedUrl.Valid {
		assert.True(t, strings.HasSuffix(dbImg.StagedUrl.String, "-staged.jpg"))
	}
}
