package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/virtual-staging-ai/worker/internal/events"
	"github.com/virtual-staging-ai/worker/internal/logging"
	"github.com/virtual-staging-ai/worker/internal/queue"
	"github.com/virtual-staging-ai/worker/internal/repository"
	"github.com/virtual-staging-ai/worker/internal/staging"
)

// ImageProcessor handles image processing jobs.
type ImageProcessor struct {
	imageRepo      repository.ImageRepository
	stagingService staging.Service
	publisher      events.Publisher
}

// NewImageProcessor creates a new image processor.
func NewImageProcessor(imageRepo repository.ImageRepository, stagingService staging.Service, publisher events.Publisher) *ImageProcessor {
	return &ImageProcessor{
		imageRepo:      imageRepo,
		stagingService: stagingService,
		publisher:      publisher,
	}
}

// JobPayload represents the payload for an image processing job.
type JobPayload struct {
	ImageID     string  `json:"image_id"`
	OriginalURL string  `json:"original_url"`
	RoomType    *string `json:"room_type,omitempty"`
	Style       *string `json:"style,omitempty"`
	Seed        *int64  `json:"seed,omitempty"`
}

// ProcessJob processes a job based on its type.
func (p *ImageProcessor) ProcessJob(ctx context.Context, job *queue.Job) error {
	tracer := otel.Tracer("virtual-staging-worker/processor")
	ctx, span := tracer.Start(ctx, "processor.ProcessJob")
	if job != nil {
		span.SetAttributes(
			attribute.String("job.id", job.ID),
			attribute.String("job.type", job.Type),
		)
	}
	defer span.End()

	switch job.Type {
	case "stage:run":
		return p.processStageJob(ctx, job)
	default:
		err := fmt.Errorf("unknown job type: %s", job.Type)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
}

// processStageJob processes an image staging job.
func (p *ImageProcessor) processStageJob(ctx context.Context, job *queue.Job) error {
	log := logging.Default()
	tracer := otel.Tracer("virtual-staging-worker/processor")
	ctx, span := tracer.Start(ctx, "processor.processStageJob")
	defer span.End()

	// Parse job payload
	var payload JobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unmarshal job payload")
		return fmt.Errorf("failed to unmarshal job payload: %w", err)
	}

	// Validate required fields
	if payload.ImageID == "" {
		err := fmt.Errorf("missing required field: image_id")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if payload.OriginalURL == "" {
		err := fmt.Errorf("missing required field: original_url")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(
		attribute.String("image.id", payload.ImageID),
	)

	log.Info(ctx, fmt.Sprintf("Processing stage job for image %s", payload.ImageID))

	// Mark image as processing
	if err := p.imageRepo.SetProcessing(ctx, payload.ImageID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "set processing failed")
		log.Error(ctx, "Failed to mark image as processing", "image_id", payload.ImageID, "error", err)
		return fmt.Errorf("failed to mark image as processing: %w", err)
	}

	// Publish processing status
	if err := p.publisher.PublishJobUpdate(ctx, events.JobUpdateEvent{
		ImageID: payload.ImageID,
		Status:  "processing",
	}); err != nil {
		log.Error(ctx, "Failed to publish processing status", "image_id", payload.ImageID, "error", err)
		// Don't fail the job if SSE publish fails
	}

	// Stage the image with AI
	stagedURL, err := p.stagingService.StageImage(ctx, &staging.StagingRequest{
		ImageID:     payload.ImageID,
		OriginalURL: payload.OriginalURL,
		RoomType:    payload.RoomType,
		Style:       payload.Style,
		Seed:        payload.Seed,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "staging failed")
		log.Error(ctx, "Failed to stage image", "image_id", payload.ImageID, "error", err)

		// Mark image as error
		if setErr := p.imageRepo.SetError(ctx, payload.ImageID, err.Error()); setErr != nil {
			log.Error(ctx, "Failed to mark image as error", "image_id", payload.ImageID, "error", setErr)
		}

		// Publish error status
		if pubErr := p.publisher.PublishJobUpdate(ctx, events.JobUpdateEvent{
			ImageID: payload.ImageID,
			Status:  "error",
			Error:   err.Error(),
		}); pubErr != nil {
			log.Error(ctx, "Failed to publish error status", "image_id", payload.ImageID, "error", pubErr)
		}

		return fmt.Errorf("failed to stage image: %w", err)
	}

	log.Info(ctx, fmt.Sprintf("Successfully staged image: %s", stagedURL))

	// Mark image as ready with staged URL
	if err := p.imageRepo.SetReady(ctx, payload.ImageID, stagedURL); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "set ready failed")
		log.Error(ctx, "Failed to mark image as ready", "image_id", payload.ImageID, "error", err)
		return fmt.Errorf("failed to mark image as ready: %w", err)
	}

	// Publish ready status
	if err := p.publisher.PublishJobUpdate(ctx, events.JobUpdateEvent{
		ImageID: payload.ImageID,
		Status:  "ready",
	}); err != nil {
		log.Error(ctx, "Failed to publish ready status", "image_id", payload.ImageID, "error", err)
		// Don't fail the job if SSE publish fails
	}

	log.Info(ctx, fmt.Sprintf("Image %s processing complete", payload.ImageID))
	span.SetStatus(codes.Ok, "processing complete")

	return nil
}
