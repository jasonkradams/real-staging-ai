package image

import (
	"context"

	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out service_mock.go . Service

// Service defines the interface for image operations.
type Service interface {
	CreateImage(ctx context.Context, req *CreateImageRequest) (*Image, error)
	GetImageByID(ctx context.Context, imageID string) (*Image, error)
	GetImagesByProjectID(ctx context.Context, projectID string) ([]*Image, error)
	UpdateImageStatus(ctx context.Context, imageID string, status Status) (*Image, error)
	UpdateImageWithStagedURL(ctx context.Context, imageID string, stagedURL string) (*Image, error)
	UpdateImageWithError(ctx context.Context, imageID string, errorMsg string) (*Image, error)
	DeleteImage(ctx context.Context, imageID string) error
	convertToImage(dbImage *queries.Image) *Image
}
