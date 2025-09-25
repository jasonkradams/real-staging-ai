package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/hibiken/asynq"
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

// AsynqQueueClient is a production-ready queue client backed by Redis + asynq.
// It adapts asynq's push-based handler model into our pull-based QueueClient API
// by bridging tasks through an internal channel and result signaling.
type AsynqQueueClient struct {
	srv     *asynq.Server
	jobs    chan *Job
	mu      sync.Mutex
	results map[string]chan error
}

// NewAsynqQueueClientFromEnv initializes an Asynq-backed queue client.
// Required env: REDIS_ADDR
// Optional env: JOB_QUEUE_NAME (default: "default"), WORKER_CONCURRENCY (default: 5)
func NewAsynqQueueClientFromEnv() (*AsynqQueueClient, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil, errors.New("REDIS_ADDR not set")
	}
	queueName := os.Getenv("JOB_QUEUE_NAME")
	if queueName == "" {
		queueName = "default"
	}
	concurrency := 5
	if v := os.Getenv("WORKER_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			concurrency = n
		}
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: addr},
		asynq.Config{
			Concurrency: concurrency,
			Queues:      map[string]int{queueName: 1},
		},
	)

	c := &AsynqQueueClient{
		srv:     srv,
		jobs:    make(chan *Job, concurrency*2),
		results: make(map[string]chan error),
	}

	mux := asynq.NewServeMux()
	// Bridge all task types; specialized routing can be added as needed.
	mux.HandleFunc("*", func(ctx context.Context, t *asynq.Task) error {
		// Create a local job id to correlate completion/failure.
		jobID := fmt.Sprintf("%d", time.Now().UnixNano())
		jb := &Job{
			ID:      jobID,
			Type:    t.Type(),
			Payload: t.Payload(),
			Status:  "queued",
		}

		// Register a result channel for this job.
		resCh := make(chan error, 1)
		c.mu.Lock()
		c.results[jobID] = resCh
		c.mu.Unlock()

		// Deliver job to consumer loop.
		select {
		case c.jobs <- jb:
		case <-ctx.Done():
			c.mu.Lock()
			delete(c.results, jobID)
			c.mu.Unlock()
			return ctx.Err()
		}

		// Wait for processing result from the worker.
		select {
		case err := <-resCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	// Start the asynq server in the background.
	go func() {
		_ = srv.Run(mux)
	}()

	return c, nil
}

// GetNextJob returns the next available job if present (non-blocking).
func (c *AsynqQueueClient) GetNextJob(ctx context.Context) (*Job, error) {
	select {
	case jb := <-c.jobs:
		return jb, nil
	default:
		return nil, nil
	}
}

// MarkJobCompleted acknowledges successful processing to the asynq handler.
func (c *AsynqQueueClient) MarkJobCompleted(ctx context.Context, jobID string) error {
	c.mu.Lock()
	ch, ok := c.results[jobID]
	if ok {
		delete(c.results, jobID)
	}
	c.mu.Unlock()
	if !ok {
		return nil
	}
	select {
	case ch <- nil:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// MarkJobFailed reports failure back to the asynq handler (will retry per queue policy).
func (c *AsynqQueueClient) MarkJobFailed(ctx context.Context, jobID string, errorMsg string) error {
	c.mu.Lock()
	ch, ok := c.results[jobID]
	if ok {
		delete(c.results, jobID)
	}
	c.mu.Unlock()
	if !ok {
		return nil
	}
	err := errors.New(errorMsg)
	select {
	case ch <- err:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
