// Package job provides domain models and types for background job operations.
package job

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Status represents the processing status of a job.
type Status string

const (
	// StatusQueued indicates the job is waiting to be processed.
	StatusQueued Status = "queued"
	// StatusProcessing indicates the job is currently being processed.
	StatusProcessing Status = "processing"
	// StatusCompleted indicates the job has been successfully completed.
	StatusCompleted Status = "completed"
	// StatusFailed indicates the job failed during processing.
	StatusFailed Status = "failed"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// Type represents the type of job to be executed.
type Type string

const (
	// TypeStageImage represents an image staging job.
	TypeStageImage Type = "stage:image"
)

// String returns the string representation of the job type.
func (t Type) String() string {
	return string(t)
}

// Job represents a background job in the system.
type Job struct {
	ID         uuid.UUID       `json:"id"`
	ImageID    uuid.UUID       `json:"image_id"`
	Type       Type            `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	Status     Status          `json:"status"`
	Error      *string         `json:"error,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	StartedAt  *time.Time      `json:"started_at,omitempty"`
	FinishedAt *time.Time      `json:"finished_at,omitempty"`
}

// CreateJobRequest represents the request to create a new job.
type CreateJobRequest struct {
	ImageID uuid.UUID       `json:"image_id" validate:"required"`
	Type    Type            `json:"type" validate:"required"`
	Payload json.RawMessage `json:"payload" validate:"required"`
}
