package events

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPublisher_Success_NoRetry(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	pub := NewDefaultPublisherWithClient(rdb, Options{MaxAttempts: 3, BaseDelay: 5 * time.Millisecond, MaxDelay: 20 * time.Millisecond})
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
    // Point to a non-listening port so Publish fails immediately
    rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6390"})
    pub := NewDefaultPublisherWithClient(rdb, Options{MaxAttempts: 2, BaseDelay: 5 * time.Millisecond, MaxDelay: 10 * time.Millisecond})

    // Capture logs
    var buf bytes.Buffer
    orig := log.Writer()
    log.SetOutput(&buf)
    t.Cleanup(func() { log.SetOutput(orig) })

    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    err := pub.PublishJobUpdate(ctx, JobUpdateEvent{JobID: "j2", ImageID: "img-fail", Status: "processing"})
    require.Error(t, err)

    // Ensure at least one retry log line captured
    out := buf.String()
    assert.True(t, strings.Contains(out, "publish failed"), "expected retry log, got: %s", out)
}

func TestDefaultPublisher_BackoffBounded(t *testing.T) {
	pub := NewDefaultPublisherWithClient(nil, Options{MaxAttempts: 5, BaseDelay: 10 * time.Millisecond, MaxDelay: 25 * time.Millisecond})
	drp, ok := pub.(*defaultRedisPublisher)
	require.True(t, ok)
	// attempt 1 => 10ms, 2 => 20ms, 3 => 25ms (capped), 4 => 25ms
	assert.InDelta(t, (10 * time.Millisecond).Seconds(), drp.backoffDelay(1).Seconds(), 0.005)
	assert.InDelta(t, (20 * time.Millisecond).Seconds(), drp.backoffDelay(2).Seconds(), 0.005)
	assert.InDelta(t, (25 * time.Millisecond).Seconds(), drp.backoffDelay(3).Seconds(), 0.005)
	assert.InDelta(t, (25 * time.Millisecond).Seconds(), drp.backoffDelay(4).Seconds(), 0.005)
}
