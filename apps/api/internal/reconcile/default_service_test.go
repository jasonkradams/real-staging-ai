package reconcile

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

func TestNewDefaultService(t *testing.T) {
	// Test the NewDefaultService constructor to ensure coverage
	dbMock := &storage.DatabaseMock{}
	s3Mock := &storage.S3ServiceMock{}

	service := NewDefaultService(dbMock, s3Mock)
	assert.NotNil(t, service)
	assert.NotNil(t, service.querier)
	assert.NotNil(t, service.s3)
}

func TestReconcileService_ReconcileImages(t *testing.T) {
	testCases := []struct {
		name        string
		opts        ReconcileOptions
		setupMocks  func(*queries.QuerierMock, *storage.S3ServiceMock)
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, result *ReconcileResult)
	}{
		{
			name: "success: no images to reconcile",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return []*queries.ListImagesForReconcileRow{}, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 0, result.Checked)
				assert.Equal(t, 0, result.MissingOrig)
				assert.Equal(t, 0, result.MissingStaged)
				assert.Equal(t, 0, result.Updated)
				assert.True(t, result.DryRun)
			},
		},
		{
			name: "success: all files exist",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
				DryRun:      false,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				img := createTestImage("img-1", "queued", "http://s3.amazonaws.com/uploads/test.jpg", "")
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return []*queries.ListImagesForReconcileRow{img}, nil
				}
				s3Mock.HeadFileFunc = func(ctx context.Context, fileKey string) (interface{}, error) {
					return struct{}{}, nil // File exists
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 1, result.Checked)
				assert.Equal(t, 0, result.MissingOrig)
				assert.Equal(t, 0, result.MissingStaged)
				assert.Equal(t, 0, result.Updated)
			},
		},
		{
			name: "success: missing original file (dry run)",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				img := createTestImage("img-1", "queued", "http://s3.amazonaws.com/uploads/test.jpg", "")
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return []*queries.ListImagesForReconcileRow{img}, nil
				}
				s3Mock.HeadFileFunc = func(ctx context.Context, fileKey string) (interface{}, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 1, result.Checked)
				assert.Equal(t, 1, result.MissingOrig)
				assert.Equal(t, 0, result.MissingStaged)
				assert.Equal(t, 0, result.Updated) // Dry-run does not update
				assert.True(t, result.DryRun)
				require.Len(t, result.Examples, 1)
				assert.Contains(t, result.Examples[0].Error, "original missing")
			},
		},
		{
			name: "success: missing staged file (dry run)",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				img := createTestImage("img-1", "ready",
					"http://s3.amazonaws.com/uploads/test.jpg",
					"http://s3.amazonaws.com/uploads/test-staged.jpg")
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return []*queries.ListImagesForReconcileRow{img}, nil
				}
				s3Mock.HeadFileFunc = func(ctx context.Context, fileKey string) (interface{}, error) {
					if fileKey == "uploads/test.jpg" {
						return struct{}{}, nil // Original exists
					}
					return nil, errors.New("staged not found")
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 1, result.Checked)
				assert.Equal(t, 0, result.MissingOrig)
				assert.Equal(t, 1, result.MissingStaged)
				assert.Equal(t, 0, result.Updated) // Dry-run does not update
				assert.True(t, result.DryRun)
				require.Len(t, result.Examples, 1)
				assert.Contains(t, result.Examples[0].Error, "staged missing")
			},
		},
		{
			name: "success: missing original file (apply changes)",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
				DryRun:      false,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				img := createTestImage("img-1", "queued", "http://s3.amazonaws.com/uploads/test.jpg", "")
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return []*queries.ListImagesForReconcileRow{img}, nil
				}
				qMock.UpdateImageWithErrorFunc = func(
					ctx context.Context, arg queries.UpdateImageWithErrorParams,
				) (*queries.UpdateImageWithErrorRow, error) {
					return &queries.UpdateImageWithErrorRow{
						ID:          img.ID,
						ProjectID:   img.ProjectID,
						OriginalUrl: img.OriginalUrl,
						StagedUrl:   img.StagedUrl,
						RoomType:    img.RoomType,
						Style:       img.Style,
						Seed:        img.Seed,
						Status:      img.Status,
						Error:       arg.Error,
						CreatedAt:   img.CreatedAt,
						UpdatedAt:   img.UpdatedAt,
					}, nil
				}
				s3Mock.HeadFileFunc = func(ctx context.Context, fileKey string) (interface{}, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 1, result.Checked)
				assert.Equal(t, 1, result.MissingOrig)
				assert.Equal(t, 1, result.Updated)
				assert.False(t, result.DryRun)
			},
		},
		{
			name: "failure: invalid project_id",
			opts: ReconcileOptions{
				ProjectID: stringPtr("invalid-uuid"),
				Limit:     100,
			},
			setupMocks:  func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {},
			expectError: true,
			errorMsg:    "invalid project_id",
		},
		{
			name: "failure: invalid cursor",
			opts: ReconcileOptions{
				Cursor: stringPtr("invalid-uuid"),
				Limit:  100,
			},
			setupMocks:  func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {},
			expectError: true,
			errorMsg:    "invalid cursor",
		},
		{
			name: "failure: database query fails",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return nil, errors.New("database connection failed")
				}
			},
			expectError: true,
			errorMsg:    "failed to list images",
		},
		{
			name: "success: multiple images with mixed results",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				img1 := createTestImage("img-1", "queued", "http://s3.amazonaws.com/uploads/test1.jpg", "")
				img2 := createTestImage("img-2", "ready",
					"http://s3.amazonaws.com/uploads/test2.jpg",
					"http://s3.amazonaws.com/uploads/test2-staged.jpg")
				img3 := createTestImage("img-3", "queued", "http://s3.amazonaws.com/uploads/test3.jpg", "")
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return []*queries.ListImagesForReconcileRow{img1, img2, img3}, nil
				}
				s3Mock.HeadFileFunc = func(ctx context.Context, fileKey string) (interface{}, error) {
					if fileKey == "uploads/test1.jpg" {
						return nil, errors.New("not found") // Missing
					}
					if fileKey == "uploads/test2-staged.jpg" {
						return nil, errors.New("not found") // Missing staged
					}
					return struct{}{}, nil // Others exist
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 3, result.Checked)
				assert.Equal(t, 1, result.MissingOrig)
				assert.Equal(t, 1, result.MissingStaged)
				assert.Equal(t, 0, result.Updated) // Dry-run does not update
				assert.True(t, result.DryRun)
			},
		},
		{
			name: "success: defaults applied when limits are zero",
			opts: ReconcileOptions{
				Limit:       0,
				Concurrency: 0,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					// Verify defaults were applied
					assert.Equal(t, int32(100), arg.Limit)
					return []*queries.ListImagesForReconcileRow{}, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 0, result.Checked)
			},
		},
		{
			name: "success: invalid original URL causes error tracking",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				img := createTestImage("img-1", "queued", "ht!tp://invalid-url", "")
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return []*queries.ListImagesForReconcileRow{img}, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 1, result.Checked)
				assert.Equal(t, 1, result.MissingOrig)
				assert.Equal(t, 0, result.Updated) // Dry-run does not update
			},
		},
		{
			name: "success: invalid staged URL causes error tracking",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				img := createTestImage("img-1", "ready", "http://s3.amazonaws.com/uploads/test.jpg", "ht!tp://invalid-staged-url")
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return []*queries.ListImagesForReconcileRow{img}, nil
				}
				s3Mock.HeadFileFunc = func(ctx context.Context, fileKey string) (interface{}, error) {
					return struct{}{}, nil // Original exists
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 1, result.Checked)
				assert.Equal(t, 0, result.MissingOrig)
				assert.Equal(t, 1, result.MissingStaged)
				assert.Equal(t, 0, result.Updated) // Dry-run does not update
			},
		},
		{
			name: "success: update failure is logged but doesn't stop reconciliation",
			opts: ReconcileOptions{
				Limit:       100,
				Concurrency: 5,
				DryRun:      false,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				img := createTestImage("img-1", "queued", "http://s3.amazonaws.com/uploads/test.jpg", "")
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					return []*queries.ListImagesForReconcileRow{img}, nil
				}
				qMock.UpdateImageWithErrorFunc = func(
					ctx context.Context, arg queries.UpdateImageWithErrorParams,
				) (*queries.UpdateImageWithErrorRow, error) {
					return nil, errors.New("database update failed")
				}
				s3Mock.HeadFileFunc = func(ctx context.Context, fileKey string) (interface{}, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *ReconcileResult) {
				assert.Equal(t, 1, result.Checked)
				assert.Equal(t, 1, result.MissingOrig)
				assert.Equal(t, 0, result.Updated) // Update failed, so not incremented
			},
		},
		{
			name: "success: with project_id filter applied",
			opts: ReconcileOptions{
				ProjectID:   stringPtr("550e8400-e29b-41d4-a716-446655440000"),
				Limit:       100,
				Concurrency: 5,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					// Verify project_id filter was applied
					assert.True(t, arg.Column1.Valid)
					return []*queries.ListImagesForReconcileRow{}, nil
				}
			},
			expectError: false,
		},
		{
			name: "success: with status filter applied",
			opts: ReconcileOptions{
				Status:      stringPtr("ready"),
				Limit:       100,
				Concurrency: 5,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					// Verify status filter was applied
					assert.Equal(t, "ready", arg.Column2)
					return []*queries.ListImagesForReconcileRow{}, nil
				}
			},
			expectError: false,
		},
		{
			name: "success: with cursor applied",
			opts: ReconcileOptions{
				Cursor:      stringPtr("550e8400-e29b-41d4-a716-446655440000"),
				Limit:       100,
				Concurrency: 5,
				DryRun:      true,
			},
			setupMocks: func(qMock *queries.QuerierMock, s3Mock *storage.S3ServiceMock) {
				qMock.ListImagesForReconcileFunc = func(
					ctx context.Context, arg queries.ListImagesForReconcileParams,
				) ([]*queries.ListImagesForReconcileRow, error) {
					// Verify cursor was applied
					assert.True(t, arg.Column3.Valid)
					return []*queries.ListImagesForReconcileRow{}, nil
				}
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			qMock := &queries.QuerierMock{}
			s3Mock := &storage.S3ServiceMock{}
			tc.setupMocks(qMock, s3Mock)

			service := NewDefaultServiceWithQuerier(qMock, s3Mock)

			result, err := service.ReconcileImages(context.Background(), tc.opts)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func createTestImage(id, status, originalURL, stagedURL string) *queries.ListImagesForReconcileRow {
	uid := uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	imgID := uuid.New()

	img := &queries.ListImagesForReconcileRow{
		ID:          pgtype.UUID{Bytes: imgID, Valid: true},
		ProjectID:   pgtype.UUID{Bytes: uid, Valid: true},
		OriginalUrl: originalURL,
		Status:      queries.ImageStatus(status),
	}

	if stagedURL != "" {
		img.StagedUrl = pgtype.Text{String: stagedURL, Valid: true}
	}

	return img
}

// Tests for helper functions

func TestExtractS3Key(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expectedKey string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "success: standard S3 URL",
			url:         "https://bucket.s3.amazonaws.com/uploads/test.jpg",
			expectedKey: "uploads/test.jpg",
			expectError: false,
		},
		{
			name:        "success: S3 URL with region",
			url:         "https://mybucket.s3.us-west-2.amazonaws.com/path/to/file.png",
			expectedKey: "path/to/file.png",
			expectError: false,
		},
		{
			name:        "success: path-style URL (MinIO)",
			url:         "http://localhost:9000/bucket/uploads/test.jpg",
			expectedKey: "uploads/test.jpg",
			expectError: false,
		},
		{
			name:        "success: path-style URL with multiple path segments",
			url:         "http://minio.example.com:9000/my-bucket/a/b/c/file.jpg",
			expectedKey: "a/b/c/file.jpg",
			expectError: false,
		},
		{
			name:        "failure: invalid URL",
			url:         "ht!tp://invalid-url",
			expectError: true,
			errorMsg:    "invalid URL",
		},
		{
			name:        "failure: empty path",
			url:         "https://bucket.s3.amazonaws.com/",
			expectError: true,
			errorMsg:    "empty path in URL",
		},
		{
			name:        "failure: path-style with no key",
			url:         "http://localhost:9000/bucket",
			expectError: true,
			errorMsg:    "cannot extract key from path-style URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := extractS3Key(tc.url)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedKey, key)
			}
		})
	}
}

func TestParseUUID(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, result [16]byte)
	}{
		{
			name:        "success: valid UUID with hyphens",
			input:       "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
			validate: func(t *testing.T, result [16]byte) {
				expected := [16]byte{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00}
				assert.Equal(t, expected, result)
			},
		},
		{
			name:        "success: valid UUID without hyphens",
			input:       "550e8400e29b41d4a716446655440000",
			expectError: false,
			validate: func(t *testing.T, result [16]byte) {
				expected := [16]byte{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00}
				assert.Equal(t, expected, result)
			},
		},
		{
			name:        "failure: too short",
			input:       "550e8400-e29b-41d4",
			expectError: true,
			errorMsg:    "invalid UUID length",
		},
		{
			name:        "failure: too long",
			input:       "550e8400-e29b-41d4-a716-446655440000-extra",
			expectError: true,
			errorMsg:    "invalid UUID length",
		},
		{
			name:        "failure: empty string",
			input:       "",
			expectError: true,
			errorMsg:    "invalid UUID length",
		},
		{
			name:        "failure: invalid hex characters",
			input:       "550e8400-e29b-41d4-a716-44665544ZZZZ",
			expectError: true,
			errorMsg:    "invalid UUID hex",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseUUID(tc.input)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				require.NoError(t, err)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}
