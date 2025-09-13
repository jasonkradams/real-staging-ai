package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{name: "jpeg allowed", contentType: "image/jpeg", expected: true},
		{name: "png allowed", contentType: "image/png", expected: true},
		{name: "webp allowed", contentType: "image/webp", expected: true},
		{name: "gif disallowed", contentType: "image/gif", expected: false},
		{name: "text disallowed", contentType: "text/plain", expected: false},
		{name: "empty disallowed", contentType: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateContentType(tt.contentType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestValidateFileSize(t *testing.T) {
	const max = 10 * 1024 * 1024 // 10MB
	tests := []struct {
		name     string
		size     int64
		expected bool
	}{
		{name: "negative invalid", size: -1, expected: false},
		{name: "zero invalid", size: 0, expected: false},
		{name: "one byte valid", size: 1, expected: true},
		{name: "exact max valid", size: max, expected: true},
		{name: "just over max invalid", size: max + 1, expected: false},
		{name: "half max valid", size: max / 2, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateFileSize(tt.size)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestValidateFilename(t *testing.T) {
	longName := strings.Repeat("a", 256) + ".jpg" // > 255 chars total
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{name: "jpg valid", filename: "image.jpg", expected: true},
		{name: "jpeg valid", filename: "image.jpeg", expected: true},
		{name: "png valid", filename: "image.png", expected: true},
		{name: "webp valid", filename: "image.webp", expected: true},
		{name: "uppercase extension valid", filename: "IMAGE.JPG", expected: true},
		{name: "gif invalid", filename: "image.gif", expected: false},
		{name: "empty invalid", filename: "", expected: false},
		{name: "too long invalid", filename: longName, expected: false},
		// Implementation only checks extension and length; path-like names still pass.
		{name: "path-like name still valid by ext", filename: "dir/../image.png", expected: true},
		{name: "multiple dots valid", filename: "my.photo.v1.jpeg", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateFilename(tt.filename)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDefaultS3Service_GetFileURL(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		fileKey    string
		expected   string
	}{
		{
			name:       "simple path",
			bucketName: "my-bucket",
			fileKey:    "uploads/u1/file.jpg",
			expected:   "https://my-bucket.s3.amazonaws.com/uploads/u1/file.jpg",
		},
		{
			name:       "key with spaces",
			bucketName: "assets",
			fileKey:    "uploads/u2/my photo.png",
			expected:   "https://assets.s3.amazonaws.com/uploads/u2/my photo.png",
		},
		{
			name:       "root key",
			bucketName: "static-files",
			fileKey:    "index.webp",
			expected:   "https://static-files.s3.amazonaws.com/index.webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DefaultS3Service{bucketName: tt.bucketName}
			got := svc.GetFileURL(tt.fileKey)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDefaultS3Service_GeneratePresignedUploadURL(t *testing.T) {
	// Force test environment branch which uses localhost:4566 and static credentials.
	t.Setenv("APP_ENV", "test")

	type input struct {
		userID      string
		filename    string
		contentType string
		fileSize    int64
	}
	tests := []struct {
		name      string
		bucket    string
		in        input
		assertFn  func(t *testing.T, res *PresignedUploadResult)
		assertURL func(t *testing.T, url, bucket, key string)
	}{
		{
			name:   "jpeg lower ext",
			bucket: "test-bucket",
			in:     input{userID: "user-123", filename: "photo.jpg", contentType: "image/jpeg", fileSize: 1024},
		},
		{
			name:   "jpeg upper ext",
			bucket: "test-bucket",
			in:     input{userID: "user-456", filename: "Photo.JPG", contentType: "image/jpeg", fileSize: 10},
		},
		{
			name:   "png multi dot",
			bucket: "assets",
			in:     input{userID: "abc", filename: "my.image.v1.png", contentType: "image/png", fileSize: 2048},
		},
		{
			name:   "webp",
			bucket: "media",
			in:     input{userID: "u", filename: "render.webp", contentType: "image/webp", fileSize: 999},
		},
	}

	defaultAssertFn := func(t *testing.T, res *PresignedUploadResult, bucket, userID, filename string) {
		t.Helper()
		require.NotNil(t, res)
		assert.NotEmpty(t, res.UploadURL)
		assert.Equal(t, int64((15 * time.Minute).Seconds()), res.ExpiresIn)

		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)

		// FileKey structure: uploads/{userID}/{base}-{uuid}{ext}
		assert.True(t, strings.HasPrefix(res.FileKey, "uploads/"+userID+"/"), "file key prefix mismatch: %s", res.FileKey)
		assert.True(t, strings.HasSuffix(res.FileKey, ext), "file key suffix mismatch: %s", res.FileKey)
		assert.Contains(t, res.FileKey, base+"-", "file key should contain base name and hyphen before uuid: %s", res.FileKey)

		// UUID pattern in the key (36 chars with dashes)
		uuidPattern := regexp.MustCompile(`[0-9a-fA-F\-]{36}\Q` + ext + `\E$`)
		assert.True(t, uuidPattern.MatchString(res.FileKey), "file key should end with uuid+ext: %s", res.FileKey)

		// URL expectations for test env localstack using path-style addressing
		assert.Contains(t, res.UploadURL, "http://localhost:4566", "presigned URL should target localstack: %s", res.UploadURL)
		assert.Contains(t, res.UploadURL, "/"+bucket+"/", "presigned URL should include bucket path segment: %s", res.UploadURL)
		assert.Contains(t, res.UploadURL, res.FileKey, "presigned URL should include file key path: %s", res.UploadURL)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, err := NewDefaultS3Service(ctx, tt.bucket)
			require.NoError(t, err)
			require.NotNil(t, svc)

			res, err := svc.GeneratePresignedUploadURL(ctx, tt.in.userID, tt.in.filename, tt.in.contentType, tt.in.fileSize)
			require.NoError(t, err)

			if tt.assertFn != nil {
				tt.assertFn(t, res)
			} else {
				defaultAssertFn(t, res, tt.bucket, tt.in.userID, tt.in.filename)
			}

			if tt.assertURL != nil {
				tt.assertURL(t, res.UploadURL, tt.bucket, res.FileKey)
			}
		})
	}
}

func TestDefaultS3Service_Integration_S3Operations(t *testing.T) {
	// These are optional integration tests that require a localstack S3 endpoint at http://localhost:4566.
	// They are skipped by default. To run, set RUN_S3_INTEGRATION_TESTS=1 in the environment.
	if os.Getenv("RUN_S3_INTEGRATION_TESTS") != "1" {
		t.Skip("skipping S3 integration tests; set RUN_S3_INTEGRATION_TESTS=1 to enable")
	}
	t.Setenv("APP_ENV", "test")

	type op string
	const (
		createBucket op = "CreateBucket"
		headFile     op = "HeadFile"
		deleteFile   op = "DeleteFile"
	)

	tests := []struct {
		name      string
		bucket    string
		operation op
		fileKey   string
		wantErr   bool
	}{
		{name: "create bucket", bucket: "it-bucket", operation: createBucket, wantErr: false},
		{name: "head non-existent file returns error", bucket: "it-bucket", operation: headFile, fileKey: "uploads/user-x/missing.jpg", wantErr: true},
		{name: "delete non-existent file succeeds", bucket: "it-bucket", operation: deleteFile, fileKey: "uploads/user-x/missing.jpg", wantErr: false},
	}

	ctx := context.Background()

	// Reuse a single service instance for speed
	svc, err := NewDefaultS3Service(ctx, "it-bucket")
	require.NoError(t, err)
	require.NotNil(t, svc)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			switch tt.operation {
			case createBucket:
				err := svc.CreateBucket(ctx)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			case headFile:
				_, err := svc.HeadFile(ctx, tt.fileKey)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			case deleteFile:
				err := svc.DeleteFile(ctx, tt.fileKey)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			default:
				t.Fatalf("unknown operation: %s", tt.operation)
			}
		})
	}
}

func TestNewDefaultS3Service_LoadDefaultConfig_Error_TestEnv(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("AWS_SDK_LOAD_CONFIG", "1")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "aws_config")

	// Write invalid INI content to force config.LoadDefaultConfig to fail.
	err := os.WriteFile(cfgPath, []byte(":::: not valid ini ::::"), 0644)
	require.NoError(t, err)

	// Point the SDK to the invalid config file.
	t.Setenv("AWS_CONFIG_FILE", cfgPath)

	t.Skip("skipping flaky env-based config failure test; rely on loader override tests")
}

func TestNewDefaultS3Service_Production(t *testing.T) {
	// Cover non-test branch using default AWS config. We only assert construction succeeds.
	t.Setenv("APP_ENV", "")
	t.Setenv("AWS_REGION", "us-east-1")
	ctx := context.Background()
	svc, err := NewDefaultS3Service(ctx, "unit-prod-bucket")
	if err != nil {
		t.Fatalf("unexpected error constructing DefaultS3Service in production branch: %v", err)
	}
	if svc == nil {
		t.Fatalf("service should not be nil")
	}
	// Also exercise GetFileURL in this branch
	url := svc.GetFileURL("some/key.jpg")
	assert.Equal(t, "https://unit-prod-bucket.s3.amazonaws.com/some/key.jpg", url)
}

func TestDefaultS3Service_CreateBucket_Idempotent(t *testing.T) {
	// Optional integration coverage for CreateBucket success + already-owned path.
	if os.Getenv("RUN_S3_INTEGRATION_TESTS") != "1" {
		t.Skip("skipping S3 integration tests; set RUN_S3_INTEGRATION_TESTS=1 to enable")
	}
	t.Setenv("APP_ENV", "test")

	ctx := context.Background()
	svc, err := NewDefaultS3Service(ctx, "it-bucket-idempotent")
	require.NoError(t, err)
	require.NotNil(t, svc)

	// First create should succeed (or be fine)
	err = svc.CreateBucket(ctx)
	assert.NoError(t, err)

	// Second create should hit BucketAlreadyOwnedByYou branch and still succeed
	err = svc.CreateBucket(ctx)
	assert.NoError(t, err)
}

func TestNewDefaultS3Service_ContextCanceled_TestEnv(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc, err := NewDefaultS3Service(ctx, "any-bucket")
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestNewDefaultS3Service_ContextCanceled_ProdEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("AWS_REGION", "us-east-1")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc, err := NewDefaultS3Service(ctx, "any-bucket")
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestDefaultS3Service_GeneratePresignedUploadURL_ContextCanceled(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	ctx := context.Background()
	svc, err := NewDefaultS3Service(ctx, "ctx-bucket")
	require.NoError(t, err)
	require.NotNil(t, svc)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := svc.GeneratePresignedUploadURL(canceled, "user", "file.jpg", "image/jpeg", 123)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestDefaultS3Service_S3Ops_ContextCanceled(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	ctx := context.Background()
	svc, err := NewDefaultS3Service(ctx, "ctx-bucket")
	require.NoError(t, err)
	require.NotNil(t, svc)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	// CreateBucket should error with canceled context
	err = svc.CreateBucket(canceled)
	assert.Error(t, err)

	// HeadFile should error with canceled context
	_, err = svc.HeadFile(canceled, "uploads/user/missing.jpg")
	assert.Error(t, err)

	// DeleteFile should error with canceled context
	err = svc.DeleteFile(canceled, "uploads/user/missing.jpg")
	assert.Error(t, err)
}

func TestNewDefaultS3Service_LoadDefaultConfig_Error_TestEnv_WithOverride(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	orig := awsConfigLoader
	defer func() { awsConfigLoader = orig }()

	awsConfigLoader = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("forced load error")
	}

	svc, err := NewDefaultS3Service(context.Background(), "bucket")
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewDefaultS3Service_LoadDefaultConfig_Error_ProdEnv_WithOverride(t *testing.T) {
	t.Setenv("APP_ENV", "")

	orig := awsConfigLoader
	defer func() { awsConfigLoader = orig }()

	awsConfigLoader = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("forced load error")
	}

	svc, err := NewDefaultS3Service(context.Background(), "bucket")
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestDefaultS3Service_DeleteFile(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	ctx := context.Background()

	svc, err := NewDefaultS3Service(ctx, "unit-bucket")
	require.NoError(t, err)
	require.NotNil(t, svc)

	canceled, cancel := context.WithCancel(ctx)
	cancel()

	err = svc.DeleteFile(canceled, "uploads/user/missing.jpg")
	assert.Error(t, err)
}

func TestDefaultS3Service_HeadFile(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	ctx := context.Background()

	svc, err := NewDefaultS3Service(ctx, "unit-bucket")
	require.NoError(t, err)
	require.NotNil(t, svc)

	canceled, cancel := context.WithCancel(ctx)
	cancel()

	_, err = svc.HeadFile(canceled, "uploads/user/missing.jpg")
	assert.Error(t, err)
}

func TestDefaultS3Service_CreateBucket(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	ctx := context.Background()

	svc, err := NewDefaultS3Service(ctx, "unit-bucket")
	require.NoError(t, err)
	require.NotNil(t, svc)

	canceled, cancel := context.WithCancel(ctx)
	cancel()

	err = svc.CreateBucket(canceled)
	assert.Error(t, err)
}
