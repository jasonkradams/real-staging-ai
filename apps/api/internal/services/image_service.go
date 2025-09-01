package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/virtual-staging-ai/api/internal/image"
	"github.com/virtual-staging-ai/api/internal/job"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// ImageService handles business logic for image operations.
type ImageService struct {
	imageRepo storage.ImageRepository
	jobRepo   storage.JobRepository
}

// NewImageService creates a new ImageService instance.
func NewImageService(imageRepo storage.ImageRepository, jobRepo storage.JobRepository) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		jobRepo:   jobRepo,
	}
}

// CreateImage creates a new image and associated job for processing.
func (s *ImageService) CreateImage(ctx context.Context, req *image.CreateImageRequest) (*image.Image, error) {
	// Validate request
	if req.ProjectID == uuid.Nil {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.OriginalURL == "" {
		return nil, fmt.Errorf("original URL is required")
	}

	// Create the image record
	imageRecord, err := s.imageRepo.CreateImage(
		ctx,
		req.ProjectID.String(),
		req.OriginalURL,
		req.RoomType,
		req.Style,
		req.Seed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	// Convert to domain model
	img := convertImageFromDB(imageRecord)

	// Create job payload
	payload := image.JobPayload{
		ImageID:     img.ID,
		OriginalURL: req.OriginalURL,
		RoomType:    req.RoomType,
		Style:       req.Style,
		Seed:        req.Seed,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job payload: %w", err)
	}

	// Create the processing job
	_, err = s.jobRepo.CreateJob(
		ctx,
		img.ID.String(),
		string(job.TypeStageImage),
		payloadJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return img, nil
}

// GetImageByID retrieves an image by its ID.
func (s *ImageService) GetImageByID(ctx context.Context, imageID string) (*image.Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return nil, fmt.Errorf("invalid image ID format: %w", err)
	}

	imageRecord, err := s.imageRepo.GetImageByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return convertImageFromDB(imageRecord), nil
}

// GetImagesByProjectID retrieves all images for a specific project.
func (s *ImageService) GetImagesByProjectID(ctx context.Context, projectID string) ([]*image.Image, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return nil, fmt.Errorf("invalid project ID format: %w", err)
	}

	imageRecords, err := s.imageRepo.GetImagesByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}

	images := make([]*image.Image, len(imageRecords))
	for i, record := range imageRecords {
		images[i] = convertImageFromDB(record)
	}

	return images, nil
}

// UpdateImageStatus updates the status of an image.
func (s *ImageService) UpdateImageStatus(ctx context.Context, imageID string, status image.Status) (*image.Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return nil, fmt.Errorf("invalid image ID format: %w", err)
	}

	imageRecord, err := s.imageRepo.UpdateImageStatus(ctx, imageID, status.String())
	if err != nil {
		return nil, fmt.Errorf("failed to update image status: %w", err)
	}

	return convertImageFromDB(imageRecord), nil
}

// UpdateImageWithStagedURL updates an image with the staged URL and marks it as ready.
func (s *ImageService) UpdateImageWithStagedURL(ctx context.Context, imageID string, stagedURL string) (*image.Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID is required")
	}
	if stagedURL == "" {
		return nil, fmt.Errorf("staged URL is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return nil, fmt.Errorf("invalid image ID format: %w", err)
	}

	imageRecord, err := s.imageRepo.UpdateImageWithStagedURL(ctx, imageID, stagedURL, string(image.StatusReady))
	if err != nil {
		return nil, fmt.Errorf("failed to update image with staged URL: %w", err)
	}

	return convertImageFromDB(imageRecord), nil
}

// UpdateImageWithError updates an image with an error status and message.
func (s *ImageService) UpdateImageWithError(ctx context.Context, imageID string, errorMsg string) (*image.Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID is required")
	}
	if errorMsg == "" {
		return nil, fmt.Errorf("error message is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return nil, fmt.Errorf("invalid image ID format: %w", err)
	}

	imageRecord, err := s.imageRepo.UpdateImageWithError(ctx, imageID, errorMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to update image with error: %w", err)
	}

	return convertImageFromDB(imageRecord), nil
}

// DeleteImage deletes an image and all associated jobs.
func (s *ImageService) DeleteImage(ctx context.Context, imageID string) error {
	if imageID == "" {
		return fmt.Errorf("image ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return fmt.Errorf("invalid image ID format: %w", err)
	}

	// Delete associated jobs first
	if err := s.jobRepo.DeleteJobsByImageID(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete jobs: %w", err)
	}

	// Delete the image
	if err := s.imageRepo.DeleteImage(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// convertImageFromDB converts a database image record to a domain model.
func convertImageFromDB(record *queries.Image) *image.Image {
	img := &image.Image{
		ID:          uuid.UUID(record.ID.Bytes),
		ProjectID:   uuid.UUID(record.ProjectID.Bytes),
		OriginalURL: record.OriginalUrl,
		Status:      image.Status(record.Status),
		CreatedAt:   record.CreatedAt.Time,
		UpdatedAt:   record.UpdatedAt.Time,
	}

	if record.StagedUrl.Valid {
		img.StagedURL = &record.StagedUrl.String
	}

	if record.RoomType.Valid {
		img.RoomType = &record.RoomType.String
	}

	if record.Style.Valid {
		img.Style = &record.Style.String
	}

	if record.Seed.Valid {
		img.Seed = &record.Seed.Int64
	}

	if record.Error.Valid {
		img.Error = &record.Error.String
	}

	return img
}
