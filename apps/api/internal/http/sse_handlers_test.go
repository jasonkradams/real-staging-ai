package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-staging-ai/api/internal/image"
	"github.com/virtual-staging-ai/api/internal/storage"
)

func TestSendSSEEvent(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
	require.NoError(t, err)
	imageServiceMock := &image.ServiceMock{}
	server := &Server{s3Service: s3ServiceMock, imageService: imageServiceMock}

	// Test event with ID and event type
	event := SSEEvent{
		ID:    "test-id",
		Event: "test-event",
		Data:  map[string]string{"message": "test message"},
	}

	err = server.sendSSEEvent(c, event)

	// Assertions
	assert.NoError(t, err)
	response := rec.Body.String()
	assert.Contains(t, response, "id: test-id\n")
	assert.Contains(t, response, "event: test-event\n")
	assert.Contains(t, response, "data: {\"message\":\"test message\"}\n\n")
}

func TestSendSSEEvent_WithoutIDAndEvent(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
	require.NoError(t, err)
	imageServiceMock := &image.ServiceMock{}
	server := &Server{s3Service: s3ServiceMock, imageService: imageServiceMock}

	// Test event without ID and event type
	event := SSEEvent{
		Data: map[string]int{"count": 42},
	}

	err = server.sendSSEEvent(c, event)

	// Assertions
	assert.NoError(t, err)
	response := rec.Body.String()
	assert.NotContains(t, response, "id:")
	assert.NotContains(t, response, "event:")
	assert.Contains(t, response, "data: {\"count\":42}\n\n")
}

func TestBroadcastJobUpdate(t *testing.T) {
	// Setup
	s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
	require.NoError(t, err)
	imageServiceMock := &image.ServiceMock{}
	server := &Server{s3Service: s3ServiceMock, imageService: imageServiceMock}

	// Test with all parameters
	errorMsg := "test error"
	progress := 50

	// This should not panic (currently just a stub)
	server.BroadcastJobUpdate("job-123", "image-456", "processing", &errorMsg, &progress)

	// Test with nil parameters
	server.BroadcastJobUpdate("job-123", "image-456", "completed", nil, nil)
}

func TestEventsHandler_InitialConnection(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
	require.NoError(t, err)
	imageServiceMock := &image.ServiceMock{}
	server := &Server{s3Service: s3ServiceMock, imageService: imageServiceMock}

	// Create a channel to signal when we want to stop the handler
	done := make(chan bool)

	// Run the handler in a goroutine
	go func() {
		_ = server.eventsHandler(c)
		done <- true
	}()

	// Wait a short time then signal completion
	select {
	case <-done:
		// Handler completed
	case <-time.After(100 * time.Millisecond):
		// Timeout - this is expected for SSE handler
	}
}

func TestSSEEventStruct(t *testing.T) {
	// Test SSEEvent JSON marshaling
	event := SSEEvent{
		ID:    "test-id",
		Event: "test-event",
		Data:  map[string]string{"key": "value"},
	}

	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "test-id")
	assert.Contains(t, string(jsonData), "test-event")
}

func TestJobUpdateEventStruct(t *testing.T) {
	// Test JobUpdateEvent JSON marshaling
	event := JobUpdateEvent{
		JobID:    "job-123",
		ImageID:  "image-456",
		Status:   "processing",
		Error:    "test error",
		Progress: 75,
	}

	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "job-123")
	assert.Contains(t, string(jsonData), "image-456")
	assert.Contains(t, string(jsonData), "processing")
	assert.Contains(t, string(jsonData), "test error")
	assert.Contains(t, string(jsonData), "75")
}
