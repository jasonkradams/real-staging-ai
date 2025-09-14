package events

import (
    "context"
    "encoding/json"
    "errors"
    "os"

    redis "github.com/redis/go-redis/v9"
)

// JobUpdateEvent mirrors the API's SSE payload for job updates.
type JobUpdateEvent struct {
    JobID    string `json:"job_id"`
    ImageID  string `json:"image_id"`
    Status   string `json:"status"`
    Error    string `json:"error,omitempty"`
    Progress int    `json:"progress,omitempty"`
}

type Publisher interface {
    PublishJobUpdate(ctx context.Context, ev JobUpdateEvent) error
}

// NewDefaultPublisherFromEnv returns a Redis-backed publisher if REDIS_ADDR is set.
func NewDefaultPublisherFromEnv() (Publisher, error) {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        return nil, errors.New("REDIS_ADDR not set")
    }
    rdb := redis.NewClient(&redis.Options{Addr: addr})
    return &redisPublisher{rdb: rdb, channel: "jobs.updates"}, nil
}

type redisPublisher struct {
    rdb     *redis.Client
    channel string
}

func (p *redisPublisher) PublishJobUpdate(ctx context.Context, ev JobUpdateEvent) error {
    b, err := json.Marshal(ev)
    if err != nil {
        return err
    }
    return p.rdb.Publish(ctx, p.channel, b).Err()
}

