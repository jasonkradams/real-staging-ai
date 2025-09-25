package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out image_repository_mock.go . ImageRepository

// ImageRepository exposes write operations necessary for the worker to
// update image processing status and final staged URL.
type ImageRepository interface {
	// SetProcessing marks the image as "processing".
	SetProcessing(ctx context.Context, imageID string) error
	// SetReady marks the image as "ready" and sets the staged URL.
	SetReady(ctx context.Context, imageID string, stagedURL string) error
	// SetError marks the image as "error" and sets the error message.
	SetError(ctx context.Context, imageID string, errorMsg string) error
}

// DefaultImageRepository is a sql.DB-backed implementation using plain SQL.
type DefaultImageRepository struct {
	db *sql.DB
}

// NewImageRepository constructs a new DefaultImageRepository.
func NewImageRepository(db *sql.DB) *DefaultImageRepository {
	return &DefaultImageRepository{db: db}
}

// SetProcessing marks the image as "processing".
func (r *DefaultImageRepository) SetProcessing(ctx context.Context, imageID string) error {
	const q = `
		UPDATE images
		SET status = 'processing', updated_at = now()
		WHERE id = $1::uuid AND status IN ('queued','processing');
	`
	if _, err := r.db.ExecContext(ctx, q, imageID); err != nil {
		return fmt.Errorf("update image status to processing: %w", err)
	}
	return nil
}

// SetReady marks the image as "ready" and sets the staged URL.
// This operation is idempotent in the sense that reapplying the same values
// does not cause an error or adverse effects.
func (r *DefaultImageRepository) SetReady(ctx context.Context, imageID string, stagedURL string) error {
	if stagedURL == "" {
		return fmt.Errorf("stagedURL cannot be empty")
	}
	const q = `
		UPDATE images
		SET staged_url = $2, status = 'ready', updated_at = now()
		WHERE id = $1::uuid AND status IN ('queued','processing');
	`
	if _, err := r.db.ExecContext(ctx, q, imageID, stagedURL); err != nil {
		return fmt.Errorf("update image with staged url: %w", err)
	}
	return nil
}

// SetError marks the image as "error" and stores an error message.
func (r *DefaultImageRepository) SetError(ctx context.Context, imageID string, errorMsg string) error {
	if errorMsg == "" {
		return fmt.Errorf("error message cannot be empty")
	}
	const q = `
		UPDATE images
		SET status = 'error', error = $2, updated_at = now()
		WHERE id = $1::uuid AND status IN ('queued','processing');
	`
	if _, err := r.db.ExecContext(ctx, q, imageID, errorMsg); err != nil {
		return fmt.Errorf("update image with error: %w", err)
	}
	return nil
}
