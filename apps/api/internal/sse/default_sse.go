package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// DefaultSSE is a Redis Pub/Sub–backed implementation of SSE.
// It streams minimal, status-only job update payloads over Server-Sent Events.
type DefaultSSE struct {
	rdb        *redis.Client
	heartbeat  time.Duration
	channelFmt string
}

// NewDefaultSSEFromEnv constructs a DefaultSSE using REDIS_ADDR from the environment.
func NewDefaultSSEFromEnv(cfg Config) (*DefaultSSE, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil, errors.New("REDIS_ADDR not set")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return NewDefaultSSE(rdb, cfg), nil
}

// NewDefaultSSE initializes a DefaultSSE with an existing Redis client.
// If cfg.HeartbeatInterval is zero, a 30s default is used.
func NewDefaultSSE(rdb *redis.Client, cfg Config) *DefaultSSE {
	hb := cfg.HeartbeatInterval
	if hb <= 0 {
		hb = 30 * time.Second
	}
	return &DefaultSSE{
		rdb:        rdb,
		heartbeat:  hb,
		channelFmt: "jobs:image:%s",
	}
}

// StreamImage subscribes to a per-image channel and forwards status-only updates via SSE.
// It emits an initial "connected" event, periodic "heartbeat" events, and "job_update" events
// containing a minimal payload: {"status":"..."}.
func (d *DefaultSSE) StreamImage(ctx context.Context, w io.Writer, imageID string) error {
	if d.rdb == nil {
		return errors.New("redis client is nil")
	}
	if imageID == "" {
		return errors.New("imageID required")
	}

	channel := fmt.Sprintf(d.channelFmt, imageID)
	sub := d.rdb.Subscribe(ctx, channel)
	defer func() { _ = sub.Close() }()

	// Optionally wait for subscription to be established
	// (Receive returns a Subscription or PONG internally; ignore value).
	if err := d.awaitSubscribe(ctx, sub); err != nil {
		return fmt.Errorf("subscribe to %s: %w", channel, err)
	}

	// Initial "connected" event
	if err := writeSSE(w, EventConnected, map[string]string{"message": "Connected to image stream"}); err != nil {
		return err
	}
	flush(w)

	// Heartbeat ticker
	ticker := time.NewTicker(d.heartbeat)
	defer ticker.Stop()

	msgCh := sub.Channel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := writeSSE(w, EventHeartbeat, map[string]any{"timestamp": time.Now().Unix()}); err != nil {
				return err
			}
			flush(w)
		case msg, ok := <-msgCh:
			if !ok {
				// Subscription channel closed; treat as done.
				return nil
			}
			// Expect minimal status-only JSON payload: {"status":"..."}
			var payload struct {
				Status string `json:"status"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil || payload.Status == "" {
				// Ignore malformed payloads to keep the stream healthy.
				continue
			}
			if err := writeSSE(w, EventJobUpdate, map[string]string{"status": payload.Status}); err != nil {
				return err
			}
			flush(w)
		}
	}
}

func (d *DefaultSSE) awaitSubscribe(ctx context.Context, sub *redis.PubSub) error {
	// Receive returns when the subscription is created or on context cancellation/error.
	_, err := sub.Receive(ctx)
	return err
}

// writeSSE writes a single Server-Sent Event to w following the SSE wire format.
func writeSSE(w io.Writer, event string, data any) error {
	if event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
			return err
		}
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", b); err != nil {
		return err
	}
	return nil
}

// flush attempts to flush the writer if it implements Flusher.
func flush(w io.Writer) {
	if f, ok := w.(Flusher); ok && f != nil {
		f.Flush()
	}
}
