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
	"github.com/real-staging-ai/worker/internal/config"
	"github.com/real-staging-ai/worker/internal/logging"
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

// NewAsynqQueueClient initializes an Asynq-backed queue client.
// Required env: REDIS_ADDR
// Optional env: JOB_QUEUE_NAME (default: "default"), WORKER_CONCURRENCY (default: 5)
func NewAsynqQueueClient(cfg *config.Config) (*AsynqQueueClient, error) {
	// check if REDIS_ADDR is set or cfg.Redis.Addr
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = cfg.Redis.Addr
	}
	if addr == "" {
		return nil, errors.New("unable to determine redis address. REDIS_ADDR env var is not set and cfg.Redis.Addr is empty")
	}

	queueName := os.Getenv("JOB_QUEUE_NAME")
	if queueName == "" {
		queueName = cfg.Job.QueueName
	}

	concurrency := cfg.Job.WorkerConcurrency
	if v := os.Getenv("WORKER_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			concurrency = n
		}
	}

	logger := logging.Default()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: addr},
		asynq.Config{
			Concurrency: concurrency,
			Queues:      map[string]int{queueName: 1},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, t *asynq.Task, err error) {
				logger.Error(ctx, "asynq handler error", "type", t.Type(), "error", err)
			}),
		},
	)

	c := &AsynqQueueClient{
		srv:     srv,
		jobs:    make(chan *Job, concurrency*2),
		results: make(map[string]chan error),
	}

	mux := asynq.NewServeMux()
	// Register exact task type used by the API enqueuer.
	// Wildcards are not supported by asynq mux.
	logger.Info(context.Background(), "Registering asynq handler", "task_type", "stage:run")

	mux.HandleFunc("stage:run", func(ctx context.Context, t *asynq.Task) error {
		logger.Info(ctx, "=== ASYNQ HANDLER CALLED ===", "task_type", t.Type())

		// Create a local job id to correlate completion/failure.
		jobID := fmt.Sprintf("%d", time.Now().UnixNano())
		jb := &Job{
			ID:      jobID,
			Type:    t.Type(),
			Payload: t.Payload(),
			Status:  "queued",
		}
		resCh := make(chan error, 1)
		c.mu.Lock()
		c.results[jobID] = resCh
		c.mu.Unlock()

		// Deliver job to consumer
		logger.Info(ctx, "asynq task received, delivering to job channel",
			"task_type", t.Type(), "job_id", jobID, "channel_len", len(c.jobs))
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
			if err != nil {
				logger.Warn(ctx, "worker marked task failed", "task_type", t.Type(), "job_id", jobID, "error", err)
			} else {
				logger.Info(ctx, "worker marked task completed", "task_type", t.Type(), "job_id", jobID)
			}
			return err
		case <-ctx.Done():
			c.mu.Lock()
			delete(c.results, jobID)
			c.mu.Unlock()
			return ctx.Err()
		}
	})

	// Start the asynq server in the background.
	logger.Info(context.Background(), "starting asynq server",
		"redis_addr", addr, "queue", queueName, "concurrency", concurrency)
	go func() {
		if err := srv.Run(mux); err != nil {
			logger.Error(context.Background(), "asynq server exited", "error", err)
		}
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
