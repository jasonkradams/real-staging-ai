package queue

import (
	"context"
	"encoding/json"
)

// Job represents a processing job.
type Job struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Status  string          `json:"status"`
}

// QueueClient defines the interface for job queue operations.
type QueueClient interface {
	GetNextJob(ctx context.Context) (*Job, error)
	MarkJobCompleted(ctx context.Context, jobID string) error
	MarkJobFailed(ctx context.Context, jobID string, errorMsg string) error
}

// MockQueueClient is a mock implementation for development/testing.
type MockQueueClient struct {
	jobs []Job
}

// NewMockQueueClient creates a new mock queue client.
func NewMockQueueClient() *MockQueueClient {
	return &MockQueueClient{
		jobs: []Job{},
	}
}

// GetNextJob returns the next job from the mock queue.
func (m *MockQueueClient) GetNextJob(ctx context.Context) (*Job, error) {
	// In a real implementation, this would connect to Redis/asynq
	// For now, return nil to simulate no jobs available
	return nil, nil
}

// MarkJobCompleted marks a job as completed.
func (m *MockQueueClient) MarkJobCompleted(ctx context.Context, jobID string) error {
	// In a real implementation, this would update the job status in the database
	return nil
}

// MarkJobFailed marks a job as failed with an error message.
func (m *MockQueueClient) MarkJobFailed(ctx context.Context, jobID string, errorMsg string) error {
	// In a real implementation, this would update the job status in the database
	return nil
}
