package image

import (
	"context"

	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out repository_mock.go . Repository

// Repository defines the interface for image data access operations.
type Repository interface {
	// CreateImage creates a new image in the database.
	CreateImage(ctx context.Context, projectID string, originalURL string, roomType, style *string, seed *int64) (*queries.Image, error)

	// GetImageByID retrieves a specific image by its ID.
	GetImageByID(ctx context.Context, imageID string) (*queries.Image, error)

	// GetImagesByProjectID retrieves all images for a specific project.
	GetImagesByProjectID(ctx context.Context, projectID string) ([]*queries.Image, error)

	// UpdateImageStatus updates an image's processing status.
	UpdateImageStatus(ctx context.Context, imageID string, status string) (*queries.Image, error)

	// UpdateImageWithStagedURL updates an image with the staged URL and status.
	UpdateImageWithStagedURL(ctx context.Context, imageID string, stagedURL string, status string) (*queries.Image, error)

	// UpdateImageWithError updates an image with an error status and message.
	UpdateImageWithError(ctx context.Context, imageID string, errorMsg string) (*queries.Image, error)

	// DeleteImage deletes an image from the database.
	DeleteImage(ctx context.Context, imageID string) error

	// DeleteImagesByProjectID deletes all images for a specific project.
	DeleteImagesByProjectID(ctx context.Context, projectID string) error

	// UpdateImageCost updates cost tracking information for an image.
	UpdateImageCost(ctx context.Context, imageID string, costUSD float64, modelUsed string, processingTimeMs int, predictionID string) error

	// GetProjectCostSummary retrieves cost summary for a project.
	GetProjectCostSummary(ctx context.Context, projectID string) (*ProjectCostSummary, error)
}
