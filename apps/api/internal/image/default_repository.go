package image

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// DefaultRepository implements the Repository interface.
type DefaultRepository struct {
	db storage.Database
}

// Ensure DefaultRepository implements Repository interface.
var _ Repository = (*DefaultRepository)(nil)

// NewDefaultRepository creates a new DefaultRepository instance.
func NewDefaultRepository(db storage.Database) *DefaultRepository {
	return &DefaultRepository{db: db}
}

// CreateImage creates a new image in the database.
func (r *DefaultRepository) CreateImage(ctx context.Context, projectID string, originalURL string, roomType, style *string, seed *int64) (*queries.Image, error) {
	q := queries.New(r.db.Pool())

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	var roomTypeText, styleText pgtype.Text
	var seedInt8 pgtype.Int8

	if roomType != nil {
		roomTypeText = pgtype.Text{String: *roomType, Valid: true}
	}

	if style != nil {
		styleText = pgtype.Text{String: *style, Valid: true}
	}

	if seed != nil {
		seedInt8 = pgtype.Int8{Int64: *seed, Valid: true}
	}

	image, err := q.CreateImage(ctx, queries.CreateImageParams{
		ProjectID:   pgtype.UUID{Bytes: projectUUID, Valid: true},
		OriginalUrl: originalURL,
		RoomType:    roomTypeText,
		Style:       styleText,
		Seed:        seedInt8,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	return image, nil
}

// GetImageByID retrieves a specific image by its ID.
func (r *DefaultRepository) GetImageByID(ctx context.Context, imageID string) (*queries.Image, error) {
	q := queries.New(r.db.Pool())

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %w", err)
	}

	image, err := q.GetImageByID(ctx, pgtype.UUID{Bytes: imageUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return image, nil
}

// GetImagesByProjectID retrieves all images for a specific project.
func (r *DefaultRepository) GetImagesByProjectID(ctx context.Context, projectID string) ([]*queries.Image, error) {
	q := queries.New(r.db.Pool())

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	images, err := q.GetImagesByProjectID(ctx, pgtype.UUID{Bytes: projectUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}

	return images, nil
}

// UpdateImageStatus updates an image's processing status.
func (r *DefaultRepository) UpdateImageStatus(ctx context.Context, imageID string, status string) (*queries.Image, error) {
	q := queries.New(r.db.Pool())

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %w", err)
	}

	image, err := q.UpdateImageStatus(ctx, queries.UpdateImageStatusParams{
		ID:     pgtype.UUID{Bytes: imageUUID, Valid: true},
		Status: queries.ImageStatus(status),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update image status: %w", err)
	}

	return image, nil
}

// UpdateImageWithStagedURL updates an image with the staged URL and status.
func (r *DefaultRepository) UpdateImageWithStagedURL(ctx context.Context, imageID string, stagedURL string, status string) (*queries.Image, error) {
	q := queries.New(r.db.Pool())

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %w", err)
	}

	image, err := q.UpdateImageWithStagedURL(ctx, queries.UpdateImageWithStagedURLParams{
		ID:        pgtype.UUID{Bytes: imageUUID, Valid: true},
		StagedUrl: pgtype.Text{String: stagedURL, Valid: true},
		Status:    queries.ImageStatus(status),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update image with staged URL: %w", err)
	}

	return image, nil
}

// UpdateImageWithError updates an image with an error status and message.
func (r *DefaultRepository) UpdateImageWithError(ctx context.Context, imageID string, errorMsg string) (*queries.Image, error) {
	q := queries.New(r.db.Pool())

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %w", err)
	}

	image, err := q.UpdateImageWithError(ctx, queries.UpdateImageWithErrorParams{
		ID:    pgtype.UUID{Bytes: imageUUID, Valid: true},
		Error: pgtype.Text{String: errorMsg, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update image with error: %w", err)
	}

	return image, nil
}

// DeleteImage deletes an image from the database.
func (r *DefaultRepository) DeleteImage(ctx context.Context, imageID string) error {
	q := queries.New(r.db.Pool())

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return fmt.Errorf("invalid image ID: %w", err)
	}

	err = q.DeleteImage(ctx, pgtype.UUID{Bytes: imageUUID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// DeleteImagesByProjectID deletes all images for a specific project.
func (r *DefaultRepository) DeleteImagesByProjectID(ctx context.Context, projectID string) error {
	q := queries.New(r.db.Pool())

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	err = q.DeleteImagesByProjectID(ctx, pgtype.UUID{Bytes: projectUUID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to delete images: %w", err)
	}

	return nil
}
