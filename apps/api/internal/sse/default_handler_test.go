package sse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForHandler(t *testing.T, timeout time.Duration, cond func() bool) {
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

func TestDefaultHandler_Events_ConnectedAndUpdate(t *testing.T) {
	// Start in-memory Redis and set REDIS_ADDR for DefaultSSE
	mr := miniredis.RunT(t)
	defer mr.Close()
	t.Setenv("REDIS_ADDR", mr.Addr())

	// Build handler from env (uses DefaultSSE)
	h, err := NewDefaultHandlerFromEnv(Config{HeartbeatInterval: 50 * time.Millisecond})
	require.NoError(t, err)

	// Prepare Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?image_id=img-123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Run handler with cancelable context to end the stream
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)
	c.SetRequest(req)

	done := make(chan struct{})
	go func() {
		_ = h.Events(c)
		close(done)
	}()

	// Wait for initial "connected" event
	waitForHandler(t, 500*time.Millisecond, func() bool {
		return strings.Contains(rec.Body.String(), "event: connected")
	})

	// Publish a status update to the per-image channel
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()
	err = rdb.Publish(ctx, "jobs:image:img-123", `{"status":"processing"}`).Err()
	require.NoError(t, err)

	// Expect job_update event with status-only payload
	waitForHandler(t, 500*time.Millisecond, func() bool {
		s := rec.Body.String()
		return strings.Contains(s, "event: job_update") && strings.Contains(s, `data: {"status":"processing"}`)
	})

	// Stop the stream and ensure the goroutine exits
	cancel()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stream did not stop after cancel")
	}

	// Assertions
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDefaultHandler_Events_MissingImageID(t *testing.T) {
	// Handler with nil SSE is fine; missing image_id is validated before SSE use
	h := NewDefaultHandler(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Events(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing image_id")
}

func TestDefaultHandler_Events_NoPubSubConfigured(t *testing.T) {
	// Handler with nil SSE should 503 when image_id is present
	h := NewDefaultHandler(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?image_id=img-xyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Events(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "pubsub not configured")
}

func TestDefaultHandler_Events_MultiUpdates(t *testing.T) {
	// Start in-memory Redis and set REDIS_ADDR
	mr := miniredis.RunT(t)
	defer mr.Close()
	t.Setenv("REDIS_ADDR", mr.Addr())

	// Build handler from env (uses DefaultSSE)
	h, err := NewDefaultHandlerFromEnv(Config{HeartbeatInterval: 50 * time.Millisecond})
	require.NoError(t, err)

	// Prepare Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?image_id=img-abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Run handler with cancelable context to end the stream
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)
	c.SetRequest(req)

	done := make(chan struct{})
	go func() {
		_ = h.Events(c)
		close(done)
	}()

	// Wait for initial "connected" event
	waitForHandler(t, 500*time.Millisecond, func() bool {
		return strings.Contains(rec.Body.String(), "event: connected")
	})

	// Publish a sequence of status updates
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()
	_ = rdb.Publish(ctx, "jobs:image:img-abc", `{"status":"processing"}`).Err()
	_ = rdb.Publish(ctx, "jobs:image:img-abc", `{"status":"ready"}`).Err()
	_ = rdb.Publish(ctx, "jobs:image:img-abc", `{"status":"error"}`).Err()

	// Expect all three updates to appear in the stream
	waitForHandler(t, 1*time.Second, func() bool {
		s := rec.Body.String()
		return strings.Count(s, "event: job_update") >= 3 &&
			strings.Contains(s, `data: {"status":"processing"}`) &&
			strings.Contains(s, `data: {"status":"ready"}`) &&
			strings.Contains(s, `data: {"status":"error"}`)
	})

	// Stop the stream and ensure the goroutine exits
	cancel()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stream did not stop after cancel")
	}

	// Assertions
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, rec.Code)
}
