// Package image provides domain models and types for image staging operations.
package image

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the processing status of an image.
type Status string

const (
	// StatusQueued indicates the image is waiting to be processed.
	StatusQueued Status = "queued"
	// StatusProcessing indicates the image is currently being processed.
	StatusProcessing Status = "processing"
	// StatusReady indicates the image has been successfully processed.
	StatusReady Status = "ready"
	// StatusError indicates an error occurred during processing.
	StatusError Status = "error"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// Image represents a staging image in the system.
type Image struct {
	ID                    uuid.UUID `json:"id"`
	ProjectID             uuid.UUID `json:"project_id"`
	OriginalURL           string    `json:"original_url"`
	StagedURL             *string   `json:"staged_url,omitempty"`
	RoomType              *string   `json:"room_type,omitempty"`
	Style                 *string   `json:"style,omitempty"`
	Seed                  *int64    `json:"seed,omitempty"`
	Status                Status    `json:"status"`
	Error                 *string   `json:"error,omitempty"`
	CostUSD               *float64  `json:"cost_usd,omitempty"`
	ModelUsed             *string   `json:"model_used,omitempty"`
	ProcessingTimeMs      *int      `json:"processing_time_ms,omitempty"`
	ReplicatePredictionID *string   `json:"replicate_prediction_id,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// CreateImageRequest represents the request to create a new staging image.
type CreateImageRequest struct {
	ProjectID   uuid.UUID `json:"project_id" validate:"required"`
	OriginalURL string    `json:"original_url" validate:"required,url"`
	//nolint:lll // struct tags are long
	RoomType *string `json:"room_type,omitempty" validate:"omitempty,oneof=living_room bedroom kitchen bathroom dining_room office"`
	//nolint:lll // struct tags are long
	Style *string `json:"style,omitempty" validate:"omitempty,oneof=modern contemporary traditional industrial scandinavian"`
	Seed        *int64    `json:"seed,omitempty" validate:"omitempty,min=1,max=4294967295"`
}

// JobPayload represents the payload for image processing jobs.
type JobPayload struct {
	ImageID     uuid.UUID `json:"image_id"`
	OriginalURL string    `json:"original_url"`
	RoomType    *string   `json:"room_type,omitempty"`
	Style       *string   `json:"style,omitempty"`
	Seed        *int64    `json:"seed,omitempty"`
}

// ProjectCostSummary represents cost aggregation for a project.
type ProjectCostSummary struct {
	ProjectID    uuid.UUID `json:"project_id"`
	TotalCostUSD float64   `json:"total_cost_usd"`
	ImageCount   int       `json:"image_count"`
	AvgCostUSD   float64   `json:"avg_cost_usd"`
}

// UpdateCostRequest represents a request to update image cost information.
type UpdateCostRequest struct {
	CostUSD               float64 `json:"cost_usd"`
	ModelUsed             string  `json:"model_used"`
	ProcessingTimeMs      int     `json:"processing_time_ms"`
	ReplicatePredictionID string  `json:"replicate_prediction_id"`
}

// BatchCreateImagesRequest represents a batch request to create multiple images.
type BatchCreateImagesRequest struct {
	Images []CreateImageRequest `json:"images" validate:"required,min=1,max=50,dive"`
}

// BatchCreateImagesResponse represents the response for batch image creation.
type BatchCreateImagesResponse struct {
	Images  []*Image              `json:"images"`
	Errors  []BatchImageError     `json:"errors,omitempty"`
	Success int                   `json:"success"`
	Failed  int                   `json:"failed"`
}

// BatchImageError represents an error for a specific image in batch creation.
type BatchImageError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}
