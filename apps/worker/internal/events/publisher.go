package events

import (
	"context"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out publisher_mock.go . Publisher

// JobUpdateEvent mirrors the API's SSE payload for job updates.
type JobUpdateEvent struct {
	JobID    string `json:"job_id"`
	ImageID  string `json:"image_id"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Progress int    `json:"progress,omitempty"`
}

// Publisher publishes job update events to a pub/sub backend (Redis),
// which the API consumes to stream Server-Sent Events (SSE).
type Publisher interface {
	// PublishJobUpdate publishes a minimal status-only payload for a given image.
	PublishJobUpdate(ctx context.Context, ev JobUpdateEvent) error
}
