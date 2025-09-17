package sse

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redis "github.com/redis/go-redis/v9"
)

type bufFlusher struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	flushes int
}

func (b *bufFlusher) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *bufFlusher) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushes++
}

func (b *bufFlusher) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}

func TestDefaultSSE_StreamImage_StatusUpdate(t *testing.T) {
	// Start in-memory Redis
	mr := miniredis.RunT(t)
	defer mr.Close()

	// Create go-redis client pointing to miniredis
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()

	// Create DefaultSSE with a short heartbeat (we won't rely on it here)
	sse := NewDefaultSSE(rdb, Config{HeartbeatInterval: 50 * time.Millisecond})

	// Create a cancellable context and an output buffer that supports flushing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := &bufFlusher{}

	// Start streaming in a goroutine
	done := make(chan struct{})
	go func() {
		_ = sse.StreamImage(ctx, w, "img-123")
		close(done)
	}()

	// Wait for initial "connected" event
	waitFor(t, 500*time.Millisecond, func() bool {
		return strings.Contains(w.String(), "event: connected")
	})

	// Publish a status update to per-image channel
	channel := "jobs:image:img-123"
	if err := rdb.Publish(ctx, channel, `{"status":"processing"}`).Err(); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	// Wait for job_update event to appear
	waitFor(t, 500*time.Millisecond, func() bool {
		s := w.String()
		return strings.Contains(s, "event: job_update") && strings.Contains(s, `data: {"status":"processing"}`)
	})

	// Cancel and ensure the goroutine exits
	cancel()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stream did not stop after cancel")
	}
}

func TestDefaultSSE_StreamImage_MissingImageID(t *testing.T) {
	// Start in-memory Redis
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()

	sse := NewDefaultSSE(rdb, Config{})

	err := sse.StreamImage(context.Background(), &bufFlusher{}, "")
	if err == nil || !strings.Contains(err.Error(), "imageID required") {
		t.Fatalf("expected imageID required error, got: %v", err)
	}
}

func TestDefaultSSE_StreamImage_IgnoresMalformedPayload(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()

	sse := NewDefaultSSE(rdb, Config{HeartbeatInterval: 100 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := &bufFlusher{}
	done := make(chan struct{})
	go func() {
		_ = sse.StreamImage(ctx, w, "img-bad")
		close(done)
	}()

	// Wait for connected event
	waitFor(t, 500*time.Millisecond, func() bool {
		return strings.Contains(w.String(), "event: connected")
	})

	// Publish malformed payloads
	channel := "jobs:image:img-bad"
	_ = rdb.Publish(ctx, channel, `{"foo":"bar"}`).Err() // missing status
	_ = rdb.Publish(ctx, channel, `not-json`).Err()

	// Ensure no job_update event appears within a small window
	time.Sleep(150 * time.Millisecond)
	if strings.Contains(w.String(), "event: job_update") {
		t.Fatalf("unexpected job_update for malformed payloads: %s", w.String())
	}

	// Cleanly stop
	cancel()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stream did not stop after cancel")
	}
}

func TestDefaultSSE_SubscribeError(t *testing.T) {
	// Start and immediately close miniredis to induce a subscribe error
	mr := miniredis.RunT(t)
	addr := mr.Addr()
	mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer func() { _ = rdb.Close() }()

	sse := NewDefaultSSE(rdb, Config{})
	err := sse.StreamImage(context.Background(), &bufFlusher{}, "img-sub-fail")
	if err == nil {
		t.Fatal("expected error due to subscription failure, got nil")
	}
}
