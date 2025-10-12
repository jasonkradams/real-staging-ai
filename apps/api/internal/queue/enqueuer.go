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

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/logging"
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

// NewAsynqEnqueuerFromEnv creates an enqueuer using environment variables.
// - REDIS_ADDR: required (e.g., "localhost:6379")
// - JOB_QUEUE_NAME: optional (defaults to "default")
func NewAsynqEnqueuerFromEnv(cfg *config.Config) (*AsynqEnqueuer, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = cfg.Redis.Addr
	}
	if addr == "" {
		return nil, errors.New(
			"redis address is not configured. " +
				"Please set REDIS_ADDR environment variable or configure in config file")
	}

	q := os.Getenv("JOB_QUEUE_NAME")
	if q == "" {
		q = cfg.Job.QueueName
	}

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: addr})
	return &AsynqEnqueuer{
		client:       client,
		defaultQueue: q,
	}, nil
}

// EnqueueStageRun enqueues a stage run job.
func (e *AsynqEnqueuer) EnqueueStageRun(
	ctx context.Context, payload StageRunPayload, opts *EnqueueOpts,
) (string, error) {
	tracer := otel.Tracer("real-staging-api/queue")
	ctx, span := tracer.Start(ctx, "queue.EnqueueStageRun")
	defer span.End()

	log := logging.NewDefaultLogger()

	// Basic validation to catch obvious mistakes early.
	if payload.ImageID == "" {
		err := errors.New("payload.image_id is required")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Error(ctx, "enqueue validation failed", "task_type", TaskTypeStageRun, "image_id", payload.ImageID, "error", err)
		return "", err
	}
	if payload.OriginalURL == "" {
		err := errors.New("payload.original_url is required")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Error(ctx, "enqueue validation failed", "task_type", TaskTypeStageRun, "image_id", payload.ImageID, "error", err)
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
		log.Error(ctx, "marshal payload failed", "task_type", TaskTypeStageRun, "image_id", payload.ImageID, "error", err)
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

	log.Info(ctx, "enqueue attempt", "task_type", TaskTypeStageRun, "image_id", payload.ImageID, "queue", selectedQueue)
	info, err := e.client.EnqueueContext(ctx, task, asynqOpts...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "enqueue error")
		log.Error(ctx, "enqueue failed",
			"task_type", TaskTypeStageRun,
			"image_id", payload.ImageID,
			"queue", selectedQueue,
			"error", err)
		return "", fmt.Errorf("enqueue stage:run: %w", err)
	}
	span.SetAttributes(
		attribute.String("queue.id", info.ID),
		attribute.String("queue.name", selectedQueue),
	)
	log.Info(ctx, "enqueued stage:run", "image_id", payload.ImageID, "queue", selectedQueue, "task_id", info.ID)
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
