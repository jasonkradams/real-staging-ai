package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// TaskTypeStageRun is the queue task type for running the staging pipeline.
const TaskTypeStageRun = "stage:run"

// StageRunPayload is the contract for a stage:run task payload.
//
// The fields align with the worker's processor expectations for Phase 1.
type StageRunPayload struct {
	ImageID     string  `json:"image_id"`
	OriginalURL string  `json:"original_url"`
	RoomType    *string `json:"room_type,omitempty"`
	Style       *string `json:"style,omitempty"`
	Seed        *int64  `json:"seed,omitempty"`
}

// EnqueueOpts controls per-task enqueue behavior (queue, retries, schedule, etc.).
// These options are intentionally generic and mapped to the underlying queue impl.
type EnqueueOpts struct {
	// Queue name to use for this task. If empty, the enqueuer's default is used.
	Queue string

	// Retry controls the maximum number of retries on failure.
	// Set to 0 for no retries. Negative value means "not set" (use queue default).
	Retry int

	// Timeout for the task execution. Zero means "not set".
	Timeout time.Duration

	// ProcessAt enqueues the task to be processed at the specified time.
	// Zero time means "not set".
	ProcessAt time.Time

	// Deadline sets the absolute deadline for the task.
	// Zero time means "not set".
	Deadline time.Time
}

// Enqueuer defines the interface for enqueuing background jobs from the API.
type Enqueuer interface {
	// EnqueueStageRun enqueues a stage:run task with the given payload.
	// Returns the task ID assigned by the queue backend.
	EnqueueStageRun(ctx context.Context, payload StageRunPayload, opts *EnqueueOpts) (string, error)
}

// AsynqEnqueuer implements Enqueuer using Redis + asynq.
type AsynqEnqueuer struct {
	client       *asynq.Client
	defaultQueue string
}

// NewAsynqEnqueuer constructs an AsynqEnqueuer with the provided Redis address
// and default queue name.
//
// Example addr: "localhost:6379"
func NewAsynqEnqueuer(addr string, defaultQueue string) (*AsynqEnqueuer, error) {
	if addr == "" {
		return nil, errors.New("redis addr is required")
	}
	if defaultQueue == "" {
		defaultQueue = "default"
	}
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: addr})
	return &AsynqEnqueuer{
		client:       client,
		defaultQueue: defaultQueue,
	}, nil
}

// NewAsynqEnqueuerFromEnv creates an enqueuer using environment variables.
// - REDIS_ADDR: required (e.g., "localhost:6379")
// - JOB_QUEUE_NAME: optional (defaults to "default")
func NewAsynqEnqueuerFromEnv() (*AsynqEnqueuer, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil, errors.New("REDIS_ADDR not set")
	}
	q := os.Getenv("JOB_QUEUE_NAME")
	if q == "" {
		q = "default"
	}
	return NewAsynqEnqueuer(addr, q)
}

// EnqueueStageRun enqueues a stage:run task with the provided payload.
func (e *AsynqEnqueuer) EnqueueStageRun(ctx context.Context, payload StageRunPayload, opts *EnqueueOpts) (string, error) {
	tracer := otel.Tracer("virtual-staging-api/queue")
	ctx, span := tracer.Start(ctx, "queue.EnqueueStageRun")
	defer span.End()

	// Basic validation to catch obvious mistakes early.
	if payload.ImageID == "" {
		err := errors.New("payload.image_id is required")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	if payload.OriginalURL == "" {
		err := errors.New("payload.original_url is required")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	span.SetAttributes(
		attribute.String("queue.task_type", TaskTypeStageRun),
		attribute.String("image.id", payload.ImageID),
	)

	b, err := json.Marshal(payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal payload")
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeStageRun, b)

	// Map our generic EnqueueOpts to asynq options.
	selectedQueue := e.defaultQueue
	asynqOpts := []asynq.Option{asynq.Queue(selectedQueue)}
	if opts != nil {
		if opts.Queue != "" {
			selectedQueue = opts.Queue
			asynqOpts[0] = asynq.Queue(selectedQueue)
		}
		if opts.Retry >= 0 {
			asynqOpts = append(asynqOpts, asynq.MaxRetry(opts.Retry))
		}
		if opts.Timeout > 0 {
			asynqOpts = append(asynqOpts, asynq.Timeout(opts.Timeout))
		}
		if !opts.ProcessAt.IsZero() {
			asynqOpts = append(asynqOpts, asynq.ProcessAt(opts.ProcessAt))
		}
		if !opts.Deadline.IsZero() {
			asynqOpts = append(asynqOpts, asynq.Deadline(opts.Deadline))
		}
	}

	info, err := e.client.EnqueueContext(ctx, task, asynqOpts...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "enqueue error")
		return "", fmt.Errorf("enqueue stage:run: %w", err)
	}
	span.SetAttributes(
		attribute.String("queue.id", info.ID),
		attribute.String("queue.name", selectedQueue),
	)
	return info.ID, nil
}

// Close releases the underlying asynq client resources.
func (e *AsynqEnqueuer) Close() error {
	return e.client.Close()
}

// NoopEnqueuer is a drop-in Enqueuer that does nothing (useful for tests).
type NoopEnqueuer struct{}

// EnqueueStageRun implements Enqueuer by returning a static ID without side effects.
func (NoopEnqueuer) EnqueueStageRun(_ context.Context, _ StageRunPayload, _ *EnqueueOpts) (string, error) {
	return "noop", nil
}
