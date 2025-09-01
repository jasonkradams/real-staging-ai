package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/virtual-staging-ai/worker/internal/queue"
)

// ImageProcessor handles image processing jobs.
type ImageProcessor struct{}

// NewImageProcessor creates a new image processor.
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
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
	switch job.Type {
	case "stage:run":
		return p.processStageJob(ctx, job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

// processStageJob processes an image staging job.
func (p *ImageProcessor) processStageJob(ctx context.Context, job *queue.Job) error {
	// Parse job payload
	var payload JobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal job payload: %w", err)
	}

	// Validate required fields
	if payload.ImageID == "" {
		return fmt.Errorf("missing required field: image_id")
	}
	if payload.OriginalURL == "" {
		return fmt.Errorf("missing required field: original_url")
	}

	log.Printf("Processing stage job for image %s", payload.ImageID)

	// Simulate image processing work
	// In Phase 1, we create a placeholder staged image
	stagedURL, err := p.createPlaceholderStagedImage(ctx, &payload)
	if err != nil {
		return fmt.Errorf("failed to create staged image: %w", err)
	}

	log.Printf("Created staged image: %s", stagedURL)

	// In a real implementation, this would:
	// 1. Update the image record in the database with the staged URL
	// 2. Mark the image status as "ready"
	// 3. Broadcast an SSE event to notify clients

	return nil
}

// createPlaceholderStagedImage creates a placeholder staged image.
// In Phase 1, this generates a mock staged URL based on the original image.
func (p *ImageProcessor) createPlaceholderStagedImage(ctx context.Context, payload *JobPayload) (string, error) {
	// Simulate processing time
	time.Sleep(2 * time.Second)

	// Generate a mock staged URL
	// In a real implementation, this would:
	// 1. Download the original image from S3
	// 2. Apply AI staging transformations
	// 3. Upload the staged image to S3
	// 4. Return the S3 URL

	stagedURL := fmt.Sprintf("%s-staged.jpg", payload.OriginalURL)

	// Add watermark simulation
	if payload.RoomType != nil {
		log.Printf("Applied room type: %s", *payload.RoomType)
	}
	if payload.Style != nil {
		log.Printf("Applied style: %s", *payload.Style)
	}
	if payload.Seed != nil {
		log.Printf("Used seed: %d", *payload.Seed)
	}

	return stagedURL, nil
}
