package http

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/labstack/echo/v4"
)

// SSEEvent represents a server-sent event.
type SSEEvent struct {
	ID    string      `json:"id,omitempty"`
	Event string      `json:"event,omitempty"`
	Data  interface{} `json:"data"`
}

// JobUpdateEvent represents a job status update event.
type JobUpdateEvent struct {
	JobID    string `json:"job_id"`
	ImageID  string `json:"image_id"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Progress int    `json:"progress,omitempty"`
}

// eventsHandler handles GET /api/v1/events requests for Server-Sent Events.
func (s *Server) eventsHandler(c echo.Context) error {
	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Send initial connection event
	if err := s.sendSSEEvent(c, SSEEvent{
		Event: "connected",
		Data: map[string]string{
			"message": "Connected to job updates stream",
		},
	}); err != nil {
		return err
	}

	// Create a context that will be cancelled when the client disconnects
	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

    // Set up heartbeat and optional Pub/Sub forwarding
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    var (
        msgCh <-chan []byte
        unsubscribe func() error
    )
    if s.pubsub != nil {
        ch, unsub, err := s.pubsub.Subscribe(ctx, "jobs.updates")
        if err == nil {
            msgCh = ch
            unsubscribe = unsub
            defer func() { if unsubscribe != nil { _ = unsubscribe() } }()
        }
    }

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            // Send heartbeat
            if err := s.sendSSEEvent(c, SSEEvent{
                Event: "heartbeat",
                Data: map[string]interface{}{
                    "timestamp": time.Now().Unix(),
                },
            }); err != nil {
                return err
            }
            c.Response().Flush()
        case data, ok := <-msgCh:
            if !ok {
                msgCh = nil
                continue
            }
            // Forward job updates (expect JSON payload)
            var ev JobUpdateEvent
            if err := json.Unmarshal(data, &ev); err == nil {
                if err := s.sendSSEEvent(c, SSEEvent{Event: "job_update", Data: ev}); err != nil {
                    return err
                }
                c.Response().Flush()
            }
        }
    }
}

// sendSSEEvent sends a Server-Sent Event to the client.
func (s *Server) sendSSEEvent(c echo.Context, event SSEEvent) error {
	// Format the event according to SSE specification
	if event.ID != "" {
		if _, err := fmt.Fprintf(c.Response(), "id: %s\n", event.ID); err != nil {
			return err
		}
	}

	if event.Event != "" {
		if _, err := fmt.Fprintf(c.Response(), "event: %s\n", event.Event); err != nil {
			return err
		}
	}

	// Marshal data to JSON
	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	if _, err := fmt.Fprintf(c.Response(), "data: %s\n\n", string(dataJSON)); err != nil {
		return err
	}

	return nil
}

// BroadcastJobUpdate broadcasts a job update to all connected SSE clients.
// In a real implementation, this would use a message broker like Redis or NATS.
func (s *Server) BroadcastJobUpdate(jobID, imageID, status string, errorMsg *string, progress *int) {
	event := JobUpdateEvent{
		JobID:   jobID,
		ImageID: imageID,
		Status:  status,
	}

	if errorMsg != nil {
		event.Error = *errorMsg
	}

	if progress != nil {
		event.Progress = *progress
	}

	// TODO: Integrate Redis Pub/Sub for broadcasting job updates to SSE clients.
	// Plan: publish JobUpdateEvent to a channel; eventsHandler subscribes and streams to connected clients.
}
