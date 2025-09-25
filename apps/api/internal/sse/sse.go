package sse

import (
	"context"
	"io"
	"time"

	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out sse_mock.go . SSE Handler

// SSE defines the contract for streaming Server-Sent Events (SSE) to clients.
//
// Implementations should:
// - Emit an initial "connected" event after a successful subscription.
// - Periodically emit "heartbeat" events at the configured interval.
// - Forward minimal, status-only job update messages received from a per-image channel.
// - Handle context cancellation for client disconnects and cleanup.
//
// The HTTP layer is responsible for setting appropriate SSE headers before
// invoking this interface, and may provide a ResponseWriter that implements
// http.Flusher. Implementations can attempt to flush by asserting the provided
// writer to a Flusher (defined below).
type SSE interface {
	// StreamImage streams events for a single image identified by imageID.
	//
	// The implementation should subscribe to a per-image channel (e.g., jobs:image:{imageID}),
	// forward status-only payloads as SSE "job_update" events, send an initial "connected"
	// event, and send periodic "heartbeat" events until ctx is cancelled.
	//
	// The writer is typically an http.ResponseWriter. If it implements Flusher,
	// the implementation should call Flush() after sending events to reduce latency.
	StreamImage(ctx context.Context, w io.Writer, imageID string) error
}

// Handler defines the HTTP-level handler for SSE endpoints, typically using Echo.
type Handler interface {
	// Events handles GET /api/v1/events?image_id={id}
	// It should set SSE headers and delegate to an SSE implementation.
	Events(c echo.Context) error
}

// Flusher is the minimal interface extracted from http.Flusher to avoid
// importing net/http in the core interface file.
type Flusher interface {
	Flush()
}

// Event names for SSE messages. Implementations may use these for consistency.
const (
	EventConnected = "connected"
	EventHeartbeat = "heartbeat"
	EventJobUpdate = "job_update"
)

// Config carries optional tuning parameters for SSE implementations.
// Implementations may choose to ignore fields if not relevant.
type Config struct {
	// HeartbeatInterval controls how often heartbeat events are emitted.
	// If zero, a reasonable default (e.g., 30s) should be used.
	HeartbeatInterval time.Duration

	// SubscribeTimeout controls how long to wait when establishing a subscription
	// to the underlying pub/sub before returning an error. If zero, implementations
	// may choose a reasonable default or rely on context deadlines.
	SubscribeTimeout time.Duration
}
