package events

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "math"
    "os"
    "time"

    redis "github.com/redis/go-redis/v9"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

// Options controls retry/backoff behavior for the default publisher.
// Zero values select sensible defaults.
type Options struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
}

// NewDefaultPublisherFromEnv returns a Redis-backed publisher if REDIS_ADDR is set.
func NewDefaultPublisherFromEnv() (Publisher, error) {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        return nil, errors.New("REDIS_ADDR not set")
    }
    rdb := redis.NewClient(&redis.Options{Addr: addr})
    return NewDefaultPublisherWithClient(rdb, Options{}), nil
}

// NewDefaultPublisherWithClient constructs a publisher with a provided redis client
// and optional retry options (for tests or custom tuning).
func NewDefaultPublisherWithClient(rdb *redis.Client, opts Options) Publisher {
    maxAttempts := opts.MaxAttempts
    if maxAttempts <= 0 {
        maxAttempts = 5
    }
    baseDelay := opts.BaseDelay
    if baseDelay <= 0 {
        baseDelay = 100 * time.Millisecond
    }
    maxDelay := opts.MaxDelay
    if maxDelay <= 0 {
        maxDelay = 2 * time.Second
    }
    return &defaultRedisPublisher{rdb: rdb, maxAttempts: maxAttempts, baseDelay: baseDelay, maxDelay: maxDelay}
}

type defaultRedisPublisher struct {
    rdb         *redis.Client
    maxAttempts int
    baseDelay   time.Duration
    maxDelay    time.Duration
}

func (p *defaultRedisPublisher) PublishJobUpdate(ctx context.Context, ev JobUpdateEvent) error {
    tracer := otel.Tracer("virtual-staging-worker/events")
    ctx, span := tracer.Start(ctx, "events.PublishJobUpdate")
    defer span.End()

    if ev.ImageID == "" {
        err := errors.New("image_id is required")
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    // Minimal payload: status only (SSE contract)
    payload, err := json.Marshal(map[string]string{"status": ev.Status})
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "marshal payload")
        return fmt.Errorf("marshal payload: %w", err)
    }
    channel := fmt.Sprintf("jobs:image:%s", ev.ImageID)
    span.SetAttributes(
        attribute.String("image.id", ev.ImageID),
        attribute.String("event.status", ev.Status),
        attribute.String("events.channel", channel),
    )

    var attempt int
    for {
        attempt++
        err = p.rdb.Publish(ctx, channel, payload).Err()
        if err == nil {
            return nil
        }
        // Log with context for observability
        log.Printf("events: publish failed image_id=%s status=%s attempt=%d err=%v", ev.ImageID, ev.Status, attempt, err)
        span.RecordError(err)
        if attempt >= p.maxAttempts {
            span.SetStatus(codes.Error, "publish attempts exceeded")
            return fmt.Errorf("publish failed after %d attempts: %w", attempt, err)
        }
        // Backoff with simple exponential increase
        delay := p.backoffDelay(attempt)
        select {
        case <-time.After(delay):
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (p *defaultRedisPublisher) backoffDelay(attempt int) time.Duration {
    // attempt starts at 1; compute delay = base * 2^(attempt-1) up to maxDelay
    exp := math.Pow(2, float64(attempt-1))
    d := time.Duration(float64(p.baseDelay) * exp)
    if d > p.maxDelay {
        return p.maxDelay
    }
    return d
}
