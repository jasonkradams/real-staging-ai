package events

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/worker/internal/logging"
)

func TestDefaultPublisher_Success_NoRetry(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	pub := NewDefaultPublisherWithClient(rdb, Options{
		MaxAttempts: 3,
		BaseDelay:   5 * time.Millisecond,
		MaxDelay:    20 * time.Millisecond,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	imageID := "img-ok"
	chName := "jobs:image:" + imageID
	sub := rdb.Subscribe(ctx, chName)
	require.NoError(t, sub.Ping(ctx))
	defer func() { _ = sub.Close() }()
	msgCh := sub.Channel()

	ev := JobUpdateEvent{JobID: "j1", ImageID: imageID, Status: "processing"}
	require.NoError(t, pub.PublishJobUpdate(ctx, ev))

	select {
	case msg := <-msgCh:
		require.NotNil(t, msg)
		assert.Equal(t, chName, msg.Channel)
		assert.JSONEq(t, `{"status":"processing"}`, msg.Payload)
	case <-ctx.Done():
		t.Fatal("timed out waiting for message")
	}
}

func TestDefaultPublisher_RetryAndFail_Logs(t *testing.T) {
	prev := logging.Default()
	memLogger := &memoryLogger{}
	logging.SetDefault(memLogger)
	t.Cleanup(func() { logging.SetDefault(prev) })
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6390"})
	pub := NewDefaultPublisherWithClient(rdb, Options{
		MaxAttempts: 2,
		BaseDelay:   5 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := pub.PublishJobUpdate(ctx, JobUpdateEvent{JobID: "j2", ImageID: "img-fail", Status: "processing"})
	require.Error(t, err)

	assert.True(t, memLogger.Contains("events publish failed"), "expected retry log, got: %v", memLogger.entries)
}

func TestDefaultPublisher_BackoffBounded(t *testing.T) {
	pub := NewDefaultPublisherWithClient(nil, Options{
		MaxAttempts: 5,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    25 * time.Millisecond,
	})
	drp, ok := pub.(*defaultRedisPublisher)
	require.True(t, ok)
	// attempt 1 => 10ms, 2 => 20ms, 3 => 25ms (capped), 4 => 25ms
	assert.InDelta(t, (10 * time.Millisecond).Seconds(), drp.backoffDelay(1).Seconds(), 0.005)
	assert.InDelta(t, (20 * time.Millisecond).Seconds(), drp.backoffDelay(2).Seconds(), 0.005)
	assert.InDelta(t, (25 * time.Millisecond).Seconds(), drp.backoffDelay(3).Seconds(), 0.005)
	assert.InDelta(t, (25 * time.Millisecond).Seconds(), drp.backoffDelay(4).Seconds(), 0.005)
}

type memoryLogger struct {
	mu      sync.Mutex
	entries []string
}

func (m *memoryLogger) record(msg string, keysAndValues ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := fmt.Sprint(append([]any{msg}, keysAndValues...)...)
	m.entries = append(m.entries, entry)
}

func (m *memoryLogger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	m.record(msg, keysAndValues...)
}
func (m *memoryLogger) Warn(ctx context.Context, msg string, keysAndValues ...any) {
	m.record(msg, keysAndValues...)
}
func (m *memoryLogger) Error(ctx context.Context, msg string, keysAndValues ...any) {
	m.record(msg, keysAndValues...)
}
func (m *memoryLogger) Debug(ctx context.Context, msg string, keysAndValues ...any) {
	m.record(msg, keysAndValues...)
}

func (m *memoryLogger) Contains(substr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.entries {
		if strings.Contains(entry, substr) {
			return true
		}
	}
	return false
}
