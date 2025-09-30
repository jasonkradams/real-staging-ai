package reconcile

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/virtual-staging-ai/api/internal/logging"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// DefaultService implements the Service interface.
type DefaultService struct {
	querier queries.Querier
	s3      storage.S3Service
}

// NewDefaultService creates a new DefaultService with a database.
func NewDefaultService(db storage.Database, s3 storage.S3Service) *DefaultService {
	return &DefaultService{
		querier: queries.New(db),
		s3:      s3,
	}
}

// NewDefaultServiceWithQuerier creates a new DefaultService with a custom querier (for testing).
func NewDefaultServiceWithQuerier(querier queries.Querier, s3 storage.S3Service) *DefaultService {
	return &DefaultService{
		querier: querier,
		s3:      s3,
	}
}

// ReconcileImages checks S3 storage for original_url and staged_url existence and updates DB accordingly.
func (s *DefaultService) ReconcileImages(ctx context.Context, opts ReconcileOptions) (*ReconcileResult, error) {
	tracer := otel.Tracer("reconcile")
	ctx, span := tracer.Start(ctx, "reconcile.images")
	defer span.End()
	span.SetAttributes(
		attribute.Bool("dry_run", opts.DryRun),
		attribute.Int("limit", opts.Limit),
		attribute.Int("concurrency", opts.Concurrency),
	)

	if opts.Limit <= 0 {
		opts.Limit = 100
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 5
	}

	// Build filter params
	params := queries.ListImagesForReconcileParams{
		Limit:   int32(opts.Limit),
		Column2: "", // Empty string means no filter (handled by SQL: $2::text = '')
	}
	if opts.ProjectID != nil {
		parsed, parseErr := parseUUID(*opts.ProjectID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid project_id: %w", parseErr)
		}
		params.Column1 = pgtype.UUID{Bytes: parsed, Valid: true}
	}
	if opts.Status != nil {
		params.Column2 = *opts.Status
	}
	if opts.Cursor != nil {
		parsed, parseErr := parseUUID(*opts.Cursor)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid cursor: %w", parseErr)
		}
		params.Column3 = pgtype.UUID{Bytes: parsed, Valid: true}
	}

	// Fetch images
	images, err := s.querier.ListImagesForReconcile(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	logger := logging.Default()
	logger.Info(ctx, "reconcile: starting image checks", "count", len(images), "dry_run", opts.DryRun)

	result := &ReconcileResult{
		Checked:  len(images),
		Examples: []ReconcileError{},
		DryRun:   opts.DryRun,
	}

	// Worker pool for concurrent S3 checks
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, opts.Concurrency)
	var mu sync.Mutex

	for _, img := range images {
		wg.Add(1)
		go func(img *queries.Image) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			checkCtx, checkSpan := tracer.Start(ctx, "reconcile.check_image")
			checkSpan.SetAttributes(attribute.String("image.id", img.ID.String()))
			defer checkSpan.End()

			// Extract key from URL
			origKey, origErr := extractS3Key(img.OriginalUrl)
			var stagedKey string
			var stagedErr error
			if img.StagedUrl.Valid {
				stagedKey, stagedErr = extractS3Key(img.StagedUrl.String)
			}

			// Check original_url
			origMissing := false
			if origErr == nil {
				_, err := s.s3.HeadFile(checkCtx, origKey)
				if err != nil {
					origMissing = true
					mu.Lock()
					result.MissingOrig++
					mu.Unlock()
				}
			} else {
				origMissing = true
				mu.Lock()
				result.MissingOrig++
				mu.Unlock()
			}

			// Check staged_url if status=ready
			stagedMissing := false
			if img.Status == "ready" && img.StagedUrl.Valid {
				if stagedErr == nil {
					_, err := s.s3.HeadFile(checkCtx, stagedKey)
					if err != nil {
						stagedMissing = true
						mu.Lock()
						result.MissingStaged++
						mu.Unlock()
					}
				} else {
					stagedMissing = true
					mu.Lock()
					result.MissingStaged++
					mu.Unlock()
				}
			}

			// Determine action
			var errorMsg string
			shouldUpdate := false

			if origMissing {
				errorMsg = "original missing in storage"
				shouldUpdate = true
			} else if stagedMissing {
				errorMsg = "staged missing in storage"
				shouldUpdate = true
			}

			if shouldUpdate {
				mu.Lock()
				if len(result.Examples) < 10 {
					result.Examples = append(result.Examples, ReconcileError{
						ImageID: img.ID.String(),
						Status:  string(img.Status),
						Error:   errorMsg,
					})
				}
				mu.Unlock()

				if !opts.DryRun {
					// Update DB to error state
					_, updateErr := s.querier.UpdateImageWithError(checkCtx, queries.UpdateImageWithErrorParams{
						ID:    img.ID,
						Error: pgtype.Text{String: errorMsg, Valid: true},
					})
					if updateErr != nil {
						logger.Warn(checkCtx, "reconcile: failed to update image", "image_id", img.ID.String(), "error", updateErr)
					} else {
						mu.Lock()
						result.Updated++
						mu.Unlock()
					}
				} else {
					mu.Lock()
					result.Updated++
					mu.Unlock()
				}
			}
		}(img)
	}

	wg.Wait()

	logger.Info(ctx, "reconcile: completed",
		"checked", result.Checked,
		"missing_original", result.MissingOrig,
		"missing_staged", result.MissingStaged,
		"updated", result.Updated,
		"dry_run", result.DryRun,
	)

	return result, nil
}

// extractS3Key extracts the object key from an S3 URL.
// Supports https://bucket.s3.region.amazonaws.com/key and http://host/bucket/key formats.
func extractS3Key(s3URL string) (string, error) {
	parsed, err := url.Parse(s3URL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	path := strings.TrimPrefix(parsed.Path, "/")
	if path == "" {
		return "", fmt.Errorf("empty path in URL")
	}

	// If hostname contains "s3", assume standard S3 URL format
	if strings.Contains(parsed.Host, "s3") {
		return path, nil
	}

	// Otherwise, assume path-style (MinIO): /bucket/key
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("cannot extract key from path-style URL")
	}
	return parts[1], nil
}

// parseUUID parses a UUID string into [16]byte.
func parseUUID(s string) ([16]byte, error) {
	var result [16]byte
	// Simple UUID parsing: strip hyphens and decode hex
	clean := strings.ReplaceAll(s, "-", "")
	if len(clean) != 32 {
		return result, fmt.Errorf("invalid UUID length")
	}
	for i := 0; i < 16; i++ {
		_, err := fmt.Sscanf(clean[i*2:i*2+2], "%02x", &result[i])
		if err != nil {
			return result, fmt.Errorf("invalid UUID hex: %w", err)
		}
	}
	return result, nil
}
