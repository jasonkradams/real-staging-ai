//go:build integration

package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MockS3Service provides a mock implementation of S3 operations for testing.
type MockS3Service struct {
	bucketName string
}

// Ensure MockS3Service implements S3Service interface.
var _ S3Service = (*MockS3Service)(nil)

// NewMockS3Service creates a new MockS3Service instance.
func NewMockS3Service(bucketName string) *MockS3Service {
	return &MockS3Service{
		bucketName: bucketName,
	}
}

// GeneratePresignedUploadURL generates a mock presigned URL for testing.
func (s *MockS3Service) GeneratePresignedUploadURL(ctx context.Context, userID, filename, contentType string, fileSize int64) (*PresignedUploadResult, error) {
	// Generate a unique file key (same logic as real S3 service)
	fileExt := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, fileExt)
	uniqueID := uuid.New().String()
	fileKey := fmt.Sprintf("uploads/%s/%s-%s%s", userID, baseName, uniqueID, fileExt)

	// Generate a mock presigned URL
	mockURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s?AWSAccessKeyId=mock&Expires=%d&Signature=mock",
		s.bucketName, fileKey, time.Now().Add(15*time.Minute).Unix())

	return &PresignedUploadResult{
		UploadURL: mockURL,
		FileKey:   fileKey,
		ExpiresIn: int64((15 * time.Minute).Seconds()),
	}, nil
}

// GetFileURL returns a mock public URL for a file.
func (s *MockS3Service) GetFileURL(fileKey string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, fileKey)
}

// DeleteFile simulates deleting a file from S3.
func (s *MockS3Service) DeleteFile(ctx context.Context, fileKey string) error {
	// In mock, we just return success
	return nil
}

// HeadFile simulates checking if a file exists in S3.
func (s *MockS3Service) HeadFile(ctx context.Context, fileKey string) (interface{}, error) {
	// Return mock metadata
	return map[string]interface{}{
		"ContentLength": int64(1024000),
		"ContentType":   "image/jpeg",
		"LastModified":  time.Now(),
	}, nil
}
