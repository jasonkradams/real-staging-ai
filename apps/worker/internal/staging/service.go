package staging

import (
	"context"
	"io"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out service_mock.go . Service

// StagingRequest contains the parameters for staging an image.
type StagingRequest struct {
	ImageID     string
	OriginalURL string
	RoomType    *string
	Style       *string
	Seed        *int64
}

// Service defines the interface for AI-powered virtual staging operations.
type Service interface {
	// StageImage processes an image with AI staging and returns the staged image URL in S3.
	// It downloads the original from S3, sends it to Replicate for processing,
	// and uploads the result back to S3.
	StageImage(ctx context.Context, req *StagingRequest) (string, error)

	// DownloadFromS3 downloads a file from S3 and returns its content.
	DownloadFromS3(ctx context.Context, fileKey string) (io.ReadCloser, error)

	// UploadToS3 uploads a file to S3 and returns the public URL.
	UploadToS3(ctx context.Context, imageID string, content io.Reader, contentType string) (string, error)
}
