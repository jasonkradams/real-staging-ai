//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/virtual-staging-ai/worker/internal/processor"
	"github.com/virtual-staging-ai/worker/internal/queue"
)

func TestProcessor_ProcessStageRunTask(t *testing.T) {
	// Setup
	ctx := context.Background()
	proc := processor.NewImageProcessor()

	// Create test job payload
	payload := map[string]interface{}{
		"image_id":     "test-image-123",
		"original_url": "https://example.com/test.jpg",
		"room_type":    "living_room",
		"style":        "modern",
		"seed":         42,
	}

	payloadJSON, err := json.Marshal(payload)
	assert.NoError(t, err)

	// Create test job
	job := &queue.Job{
		ID:      "test-job-123",
		Type:    "stage:run",
		Payload: payloadJSON,
		Status:  "pending",
	}

	// Process the job
	err = proc.ProcessJob(ctx, job)

	// Assertions
	assert.NoError(t, err)
	// Note: In the current implementation, job status is not automatically updated
	// This would be handled by the queue client in a real implementation
}

func TestProcessor_ProcessInvalidTask(t *testing.T) {
	// Setup
	ctx := context.Background()
	proc := processor.NewImageProcessor()

	// Create test job with invalid type
	job := &queue.Job{
		ID:      "test-job-123",
		Type:    "invalid:task",
		Payload: []byte(`{}`),
		Status:  "pending",
	}

	// Process the job
	err := proc.ProcessJob(ctx, job)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown job type")
}

func TestProcessor_ProcessStageRunWithInvalidPayload(t *testing.T) {
	// Setup
	ctx := context.Background()
	proc := processor.NewImageProcessor()

	// Create test job with invalid payload
	job := &queue.Job{
		ID:      "test-job-123",
		Type:    "stage:run",
		Payload: []byte(`{"invalid": "payload"}`),
		Status:  "pending",
	}

	// Process the job
	err := proc.ProcessJob(ctx, job)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field")
}
