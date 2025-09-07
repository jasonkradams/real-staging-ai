package image

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/virtual-staging-ai/api/internal/job"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

var jsonMarshal = json.Marshal

// DefaultService handles business logic for image operations.
type DefaultService struct {
	imageRepo Repository
	jobRepo   job.Repository
}

// NewDefaultService creates a new DefaultService instance.
func NewDefaultService(imageRepo Repository, jobRepo job.Repository) *DefaultService {
	return &DefaultService{
		imageRepo: imageRepo,
		jobRepo:   jobRepo,
	}
}

// CreateImage creates a new image and queues it for processing.
func (s *DefaultService) CreateImage(ctx context.Context, req *CreateImageRequest) (*Image, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Create the image in the database
	dbImage, err := s.imageRepo.CreateImage(
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

	// Convert database image to domain image
	domainImage := s.convertToImage(dbImage)

	// Create job payload
	payload := JobPayload{
		ImageID:     domainImage.ID,
		OriginalURL: domainImage.OriginalURL,
		RoomType:    domainImage.RoomType,
		Style:       domainImage.Style,
		Seed:        domainImage.Seed,
	}

	payloadJSON, err := jsonMarshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job payload: %w", err)
	}

	// Create a job for processing the image
	_, err = s.jobRepo.CreateJob(ctx, domainImage.ID.String(), "stage:run", payloadJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return domainImage, nil
}

// GetImageByID retrieves a specific image by its ID.
func (s *DefaultService) GetImageByID(ctx context.Context, imageID string) (*Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID cannot be empty")
	}

	dbImage, err := s.imageRepo.GetImageByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return s.convertToImage(dbImage), nil
}

// GetImagesByProjectID retrieves all images for a specific project.
func (s *DefaultService) GetImagesByProjectID(ctx context.Context, projectID string) ([]*Image, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	dbImages, err := s.imageRepo.GetImagesByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}

	images := make([]*Image, len(dbImages))
	for i, dbImage := range dbImages {
		images[i] = s.convertToImage(dbImage)
	}

	return images, nil
}

// UpdateImageStatus updates an image's processing status.
func (s *DefaultService) UpdateImageStatus(ctx context.Context, imageID string, status Status) (*Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID cannot be empty")
	}

	dbImage, err := s.imageRepo.UpdateImageStatus(ctx, imageID, status.String())
	if err != nil {
		return nil, fmt.Errorf("failed to update image status: %w", err)
	}

	return s.convertToImage(dbImage), nil
}

// UpdateImageWithStagedURL updates an image with the staged URL and status.
func (s *DefaultService) UpdateImageWithStagedURL(ctx context.Context, imageID string, stagedURL string) (*Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID cannot be empty")
	}
	if stagedURL == "" {
		return nil, fmt.Errorf("staged URL cannot be empty")
	}

	dbImage, err := s.imageRepo.UpdateImageWithStagedURL(ctx, imageID, stagedURL, StatusReady.String())
	if err != nil {
		return nil, fmt.Errorf("failed to update image with staged URL: %w", err)
	}

	return s.convertToImage(dbImage), nil
}

// UpdateImageWithError updates an image with an error status and message.
func (s *DefaultService) UpdateImageWithError(ctx context.Context, imageID string, errorMsg string) (*Image, error) {
	if imageID == "" {
		return nil, fmt.Errorf("image ID cannot be empty")
	}
	if errorMsg == "" {
		return nil, fmt.Errorf("error message cannot be empty")
	}

	dbImage, err := s.imageRepo.UpdateImageWithError(ctx, imageID, errorMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to update image with error: %w", err)
	}

	return s.convertToImage(dbImage), nil
}

// DeleteImage deletes an image from the database.
func (s *DefaultService) DeleteImage(ctx context.Context, imageID string) error {
	if imageID == "" {
		return fmt.Errorf("image ID cannot be empty")
	}

	err := s.imageRepo.DeleteImage(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// convertToImage converts a database image to a domain image.
func (s *DefaultService) convertToImage(dbImage *queries.Image) *Image {
	image := &Image{
		ID:          dbImage.ID.Bytes,
		ProjectID:   dbImage.ProjectID.Bytes,
		OriginalURL: dbImage.OriginalUrl,
		Status:      Status(dbImage.Status),
		CreatedAt:   dbImage.CreatedAt.Time,
		UpdatedAt:   dbImage.UpdatedAt.Time,
	}

	if dbImage.StagedUrl.Valid {
		image.StagedURL = &dbImage.StagedUrl.String
	}

	if dbImage.RoomType.Valid {
		image.RoomType = &dbImage.RoomType.String
	}

	if dbImage.Style.Valid {
		image.Style = &dbImage.Style.String
	}

	if dbImage.Seed.Valid {
		image.Seed = &dbImage.Seed.Int64
	}

	if dbImage.Error.Valid {
		image.Error = &dbImage.Error.String
	}

	return image
}
