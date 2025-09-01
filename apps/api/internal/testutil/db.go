package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/virtual-staging-ai/api/internal/services"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/storage/mocks"
	"go.uber.org/mock/gomock"
)

func TruncateTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	query := `
		TRUNCATE TABLE projects, users, images, jobs, plans RESTART IDENTITY
	`
	_, err := pool.Exec(context.Background(), query)
	require.NoError(t, err)
}

func SeedTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// This path is relative to the package that is being tested.
	// This is not ideal, but it's the simplest solution for now.
	seedSQL, err := os.ReadFile("../../testdata/seed.sql")
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), string(seedSQL))
	require.NoError(t, err)
}

func CreateMockS3Service(t *testing.T) storage.S3Service {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockS3 := mocks.NewMockS3Service(ctrl)

	// Set up dynamic mock behaviors for presigned upload URL
	mockS3.EXPECT().
		GeneratePresignedUploadURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, userID, filename, contentType string, fileSize int64) (*storage.PresignedUploadResult, error) {
			// Generate a file key similar to the real implementation
			fileExt := filepath.Ext(filename)
			baseName := strings.TrimSuffix(filename, fileExt)
			uniqueID := "mock-123" // Use a consistent ID for testing
			fileKey := fmt.Sprintf("uploads/%s/%s-%s%s", userID, baseName, uniqueID, fileExt)

			return &storage.PresignedUploadResult{
				UploadURL: fmt.Sprintf("https://mock-bucket.s3.amazonaws.com/%s", fileKey),
				FileKey:   fileKey,
				ExpiresIn: 900,
			}, nil
		}).
		AnyTimes()

	return mockS3
}

func CreateMockImageService(t *testing.T) *services.ImageService {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockImageRepo := mocks.NewMockImageRepository(ctrl)
	mockJobRepo := mocks.NewMockJobRepository(ctrl)

	return services.NewImageService(mockImageRepo, mockJobRepo)
}
