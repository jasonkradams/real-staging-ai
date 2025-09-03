package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// PresignedUploadResult contains the result of generating a presigned upload URL.
type PresignedUploadResult struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresIn int64  `json:"expires_in"`
}

// DefaultS3Service handles S3 operations for file storage.
type DefaultS3Service struct {
	client     *s3.Client
	bucketName string
}

// Ensure DefaultS3Service implements S3Service interface.
var _ S3Service = (*DefaultS3Service)(nil)

// NewS3Service creates a new DefaultS3Service instance.
func NewS3Service(ctx context.Context, bucketName string) (*DefaultS3Service, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &DefaultS3Service{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// GeneratePresignedUploadURL generates a presigned URL for uploading a file to S3.
func (s *DefaultS3Service) GeneratePresignedUploadURL(ctx context.Context, userID, filename, contentType string, fileSize int64) (*PresignedUploadResult, error) {
	// Generate a unique file key
	fileExt := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, fileExt)
	uniqueID := uuid.New().String()
	fileKey := fmt.Sprintf("uploads/%s/%s-%s%s", userID, baseName, uniqueID, fileExt)

	// Create the presign client
	presignClient := s3.NewPresignClient(s.client)

	// Set the expiration time (15 minutes)
	expirationDuration := 15 * time.Minute

	// Create the presign request
	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(fileKey),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(fileSize),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expirationDuration
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &PresignedUploadResult{
		UploadURL: request.URL,
		FileKey:   fileKey,
		ExpiresIn: int64(expirationDuration.Seconds()),
	}, nil
}

// GetFileURL returns the public URL for a file in S3.
func (s *DefaultS3Service) GetFileURL(fileKey string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, fileKey)
}

// DeleteFile deletes a file from S3.
func (s *DefaultS3Service) DeleteFile(ctx context.Context, fileKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// HeadFile checks if a file exists in S3 and returns its metadata.
func (s *DefaultS3Service) HeadFile(ctx context.Context, fileKey string) (interface{}, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return result, nil
}

// ValidateContentType checks if the content type is allowed for uploads.
func ValidateContentType(contentType string) bool {
	allowedTypes := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}

	return false
}

// ValidateFileSize checks if the file size is within allowed limits.
func ValidateFileSize(size int64) bool {
	const maxSize = 10 * 1024 * 1024 // 10MB
	return size > 0 && size <= maxSize
}

// ValidateFilename checks if the filename is valid.
func ValidateFilename(filename string) bool {
	if len(filename) == 0 || len(filename) > 255 {
		return false
	}

	// Check for valid file extensions
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".webp"}

	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}

	return false
}
