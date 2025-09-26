package sse

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/logging"
)

// DefaultHandler provides Echo HTTP handlers for SSE endpoints.
type DefaultHandler struct {
	sse SSE
}

// NewDefaultHandler constructs a DefaultHandler with the provided SSE implementation.
func NewDefaultHandler(s SSE) *DefaultHandler {
	return &DefaultHandler{sse: s}
}

// NewDefaultHandlerFromEnv constructs a DefaultHandler using the default Redis-backed SSE implementation.
// It reads REDIS_ADDR from the environment.
func NewDefaultHandlerFromEnv(cfg Config) (*DefaultHandler, error) {
	streamer, err := NewDefaultSSEFromEnv(cfg)
	if err != nil {
		return nil, err
	}
	return &DefaultHandler{sse: streamer}, nil
}

// Events is an Echo handler for GET /api/v1/events?image_id={id} that streams
// Server-Sent Events scoped to a single image (per-image channel).
//
// It sets the appropriate SSE headers, validates the image_id query parameter,
// and delegates streaming to the configured SSE implementation.
//
// Expected minimal payloads are status-only job updates, e.g.:
//
//	event: job_update
//	data: {"status":"processing"}
func (h *DefaultHandler) Events(c echo.Context) error {
	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	imageID := c.QueryParam("image_id")
	if imageID == "" {
		logging.NewDefaultLogger().Warn(c.Request().Context(), "missing image_id for SSE events")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing image_id"})
	}

	if h.sse == nil {
		logging.NewDefaultLogger().Error(c.Request().Context(), "pubsub not configured for SSE", "image_id", imageID)
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "pubsub not configured"})
	}

	// Stream events until client disconnects (request context is cancelled)
	return h.sse.StreamImage(c.Request().Context(), c.Response().Writer, imageID)
}
