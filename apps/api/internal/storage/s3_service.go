package storage

import "context"

//go:generate go run github.com/matryer/moq@v0.5.3 -out s3_service_mock.go . S3Service

// S3Service defines the interface for S3 storage operations.
type S3Service interface {
	// HeadFile checks if a file exists in S3 and returns its metadata.
	HeadFile(ctx context.Context, fileKey string) (interface{}, error)
	// DeleteFile deletes a file from S3.
	DeleteFile(ctx context.Context, fileKey string) error
	// GetFileURL returns the public URL for a file in S3.
	GetFileURL(fileKey string) string
	// GeneratePresignedUploadURL generates a presigned URL for uploading a file to S3.
	GeneratePresignedUploadURL(ctx context.Context, userID, filename, contentType string, fileSize int64) (*PresignedUploadResult, error)
	// CreateBucket creates the S3 bucket if it doesn't exist.
	CreateBucket(ctx context.Context) error
}
