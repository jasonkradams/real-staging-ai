package integration

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/storage"
)

func TestS3Integration_UploadHeadDelete(t *testing.T) {
	if os.Getenv("RUN_S3_INTEGRATION_TESTS") != "1" {
		t.Skip("skipping S3 integration tests; set RUN_S3_INTEGRATION_TESTS=1 to enable")
	}
	// Ensure we use the Localstack branch in NewDefaultS3Service
	t.Setenv("APP_ENV", "test")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const (
		bucket      = "vsa-s3-it-bucket"
		userID      = "integration-user"
		filename    = "placeholder.png"
		contentType = "image/png"
	)

	// Load a small image from testdata as our upload payload
	dataPath := filepath.Join("testdata", "placeholder.png")
	// #nosec G304 -- Reading from fixed testdata directory in tests
	fileBytes, err := os.ReadFile(dataPath)
	require.NoError(t, err, "failed to read test image: %s", dataPath)

	svc, err := storage.NewDefaultS3Service(ctx, &config.S3{
		BucketName: bucket,
	})
	require.NoError(t, err)

	// Create bucket (idempotent)
	err = svc.CreateBucket(ctx)
	require.NoError(t, err)

	// Generate a presigned URL for upload
	presigned, err := svc.GeneratePresignedUploadURL(ctx, userID, filename, contentType, int64(len(fileBytes)))
	require.NoError(t, err)
	require.NotNil(t, presigned)
	require.NotEmpty(t, presigned.UploadURL)
	require.NotEmpty(t, presigned.FileKey)

	// PUT the file using the presigned URL; ensure Content-Type matches the presign
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, presigned.UploadURL, bytes.NewReader(fileBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", contentType)

	httpClient := &http.Client{Timeout: 20 * time.Second}
	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	// apps/api/tests/integration/storage_s3_service.go:66:24: Error return value of `io.Copy` is not checked (errcheck)
	_, _ = io.Copy(io.Discard, resp.Body)
	defer func() { _ = resp.Body.Close() }()

	// Localstack typically returns 200 for successful PUT
	require.Truef(t, resp.StatusCode >= 200 && resp.StatusCode < 300, "unexpected status from PUT: %d", resp.StatusCode)

	// HEAD the uploaded object; verify metadata
	headRes, err := svc.HeadFile(ctx, presigned.FileKey)
	require.NoError(t, err)
	require.NotNil(t, headRes)

	if out, ok := headRes.(*s3.HeadObjectOutput); ok {
		// ContentType and ContentLength should match what we uploaded
		if out.ContentType != nil {
			assert.Equal(t, contentType, *out.ContentType)
		}
		assert.Equal(t, int64(len(fileBytes)), out.ContentLength)
	}

	// DELETE the object
	err = svc.DeleteFile(ctx, presigned.FileKey)
	require.NoError(t, err)

	// Verify the object is gone; HEAD should fail now
	_, err = svc.HeadFile(ctx, presigned.FileKey)
	assert.Error(t, err, "expected error when heading deleted object")
}
