package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"

	configLib "github.com/real-staging-ai/api/internal/config"
)

// PresignedUploadResult contains the result of generating a presigned upload URL.
type PresignedUploadResult struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresIn int64  `json:"expires_in"`
}

// GeneratePresignedGetURL generates a browser-accessible presigned GET URL for a specific file key.
func (s *DefaultS3Service) GeneratePresignedGetURL(ctx context.Context, fileKey string, expiresInSeconds int64, contentDisposition string) (string, error) {
	// Choose a client for presigning that uses the public endpoint if set.
	presignBase := s.client

	// Use stored config for presign operations
	if s.Cfg != nil && s.Cfg.PublicEndpoint != "" {
		publicEndpoint := s.Cfg.PublicEndpoint
		region := s.Cfg.Region
		accessKey := s.Cfg.AccessKey
		secretKey := s.Cfg.SecretKey
		usePathStyle := s.Cfg.UsePathStyle
		presignCfg, cfgErr := awsConfigLoader(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if cfgErr == nil {
			presignBase = s3.NewFromConfig(presignCfg, func(o *s3.Options) {
				o.BaseEndpoint = aws.String(publicEndpoint)
				o.UsePathStyle = usePathStyle
			})
		}
	}

	presignClient := s3.NewPresignClient(presignBase)
	exp := time.Duration(expiresInSeconds) * time.Second
	if exp <= 0 {
		exp = 10 * time.Minute
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.Cfg.BucketName),
		Key:    aws.String(fileKey),
	}
	if contentDisposition != "" {
		input.ResponseContentDisposition = aws.String(contentDisposition)
	}
	// Let caller/browser infer content-type when omitted; optionally we could set ResponseContentType.

	req, err := presignClient.PresignGetObject(ctx, input, func(o *s3.PresignOptions) {
		o.Expires = exp
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}
	return req.URL, nil
}

// DefaultS3Service handles S3 operations for file storage.
type DefaultS3Service struct {
	client *s3.Client
	Cfg    *configLib.S3 // Store config for presign operations
}

// Ensure DefaultS3Service implements S3Service interface.
var _ S3Service = (*DefaultS3Service)(nil)

// awsConfigLoader allows overriding AWS config loading in tests.
var awsConfigLoader = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, optFns...)
}

// NewDefaultS3Service creates a new DefaultS3Service instance.
// s3Cfg must not be nil.
func NewDefaultS3Service(ctx context.Context, s3Cfg *configLib.S3) (*DefaultS3Service, error) {
	if s3Cfg == nil {
		return nil, fmt.Errorf("S3 config is required")
	}

	var cfg aws.Config
	var err error

	// Use config values
	region := s3Cfg.Region
	endpoint := s3Cfg.Endpoint
	accessKey := s3Cfg.AccessKey
	secretKey := s3Cfg.SecretKey
	usePathStyle := s3Cfg.UsePathStyle

	if os.Getenv("APP_ENV") == "test" {
		cfg, err = awsConfigLoader(ctx,
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "test")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config for test: %w", err)
		}

		client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String("http://localhost:4566")
			o.UsePathStyle = true
		})

		return &DefaultS3Service{
			client: client,
			Cfg:    s3Cfg,
		}, nil

	}

	// If a custom S3 endpoint is provided (e.g., MinIO), configure client for dev/local
	if endpoint != "" {
		cfg, err = awsConfigLoader(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config for dev: %w", err)
		}

		client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = usePathStyle
		})

		return &DefaultS3Service{
			client: client,
			Cfg:    s3Cfg,
		}, nil
	}

	// Use default AWS config for production
	cfg, err = awsConfigLoader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &DefaultS3Service{
		client: client,
		Cfg:    s3Cfg,
	}, nil
}

// GeneratePresignedUploadURL generates a presigned URL for uploading a file to S3.
func (s *DefaultS3Service) GeneratePresignedUploadURL(ctx context.Context, userID, filename, contentType string, fileSize int64) (*PresignedUploadResult, error) {
	// Generate a unique file key
	fileExt := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, fileExt)
	uniqueID := uuid.New().String()
	fileKey := fmt.Sprintf("uploads/%s/%s-%s%s", userID, baseName, uniqueID, fileExt)

	// Choose a client for presigning. If public endpoint is set, use a client
	// with that base endpoint so the URL host is browser-accessible. Provide
	// static credentials to avoid IMDS.
	presignBase := s.client

	// Use stored config for presign operations
	if s.Cfg != nil && s.Cfg.PublicEndpoint != "" {
		publicEndpoint := s.Cfg.PublicEndpoint
		region := s.Cfg.Region
		accessKey := s.Cfg.AccessKey
		secretKey := s.Cfg.SecretKey
		usePathStyle := s.Cfg.UsePathStyle
		presignCfg, cfgErr := awsConfigLoader(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if cfgErr == nil {
			presignBase = s3.NewFromConfig(presignCfg, func(o *s3.Options) {
				o.BaseEndpoint = aws.String(publicEndpoint)
				o.UsePathStyle = usePathStyle
			})
		}
	}
	// Create the presign client
	presignClient := s3.NewPresignClient(presignBase)

	// Set the expiration time (15 minutes)
	expirationDuration := 15 * time.Minute

	// Create the presign request
	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Cfg.BucketName),
		Key:         aws.String(fileKey),
		ContentType: aws.String(contentType),
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
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.Cfg.BucketName, fileKey)
}

// DeleteFile deletes a file from S3.
func (s *DefaultS3Service) DeleteFile(ctx context.Context, fileKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Cfg.BucketName),
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
		Bucket: aws.String(s.Cfg.BucketName),
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

	return slices.Contains(allowedTypes, contentType)
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

	return slices.Contains(allowedExts, ext)
}

// CreateBucket creates the S3 bucket if it doesn't exist.
func (s *DefaultS3Service) CreateBucket(ctx context.Context) error {
	_, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &s.Cfg.BucketName,
	})
	if err != nil {
		// If the bucket already exists, we can ignore the error.
		var aerr *types.BucketAlreadyOwnedByYou
		if errors.As(err, &aerr) {
			return nil
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}
