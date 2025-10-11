package events

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/worker/internal/config"
)

func TestNewDefaultPublisher_MissingEnv(t *testing.T) {
	// Ensure REDIS_ADDR is unset
	t.Setenv("REDIS_ADDR", "")

	_, err := NewDefaultPublisher(&config.Config{})
	require.Error(t, err, "expected error when REDIS_ADDR is not set")
}

func TestRedisPublisher_PublishJobUpdate_SendsStatusOnly(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	require.NoError(t, err)

	// Start in-memory Redis
	mr := miniredis.RunT(t)
	defer mr.Close()

	// Configure env for publisher
	t.Setenv("REDIS_ADDR", mr.Addr())

	// Build publisher from env
	pub, err := NewDefaultPublisher(cfg)
	require.NoError(t, err)

	// Create a Redis client for subscribing
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	imageID := "img-123"
	channel := "jobs:image:" + imageID

	// Subscribe to the per-image channel
	sub := rdb.Subscribe(ctx, channel)
	defer func() { _ = sub.Close() }()

	// Ensure subscription is active
	_, err = sub.Receive(ctx)
	require.NoError(t, err, "failed to establish subscription")

	// Start a goroutine to receive the message
	msgCh := make(chan *redis.Message, 1)
	go func() {
		defer close(msgCh)
		msg, _ := sub.ReceiveMessage(ctx)
		msgCh <- msg
	}()

	// Publish via the events publisher
	ev := JobUpdateEvent{
		JobID:   "job-1",
		ImageID: imageID,
		Status:  "processing",
		// Note: Error and other fields are intentionally omitted in payload
	}
	err = pub.PublishJobUpdate(ctx, ev)
	require.NoError(t, err)

	// Assert we received the expected message on the expected channel
	select {
	case msg := <-msgCh:
		require.NotNil(t, msg, "expected a message")
		assert.Equal(t, channel, msg.Channel)
		// Publisher only emits {"status": "<value>"} by contract
		assert.JSONEq(t, `{"status":"processing"}`, msg.Payload)
	case <-ctx.Done():
		t.Fatal("timed out waiting for pubsub message")
	}
}

func TestRedisPublisher_PublishJobUpdate_DifferentImageChannel(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	require.NoError(t, err)

	mr := miniredis.RunT(t)
	defer mr.Close()

	t.Setenv("REDIS_ADDR", mr.Addr())
	pub, err := NewDefaultPublisher(cfg)
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	imageID := "another-image"
	channel := "jobs:image:" + imageID

	sub := rdb.Subscribe(ctx, channel)
	defer func() { _ = sub.Close() }()

	_, err = sub.Receive(ctx)
	require.NoError(t, err)

	msgCh := make(chan *redis.Message, 1)
	go func() {
		defer close(msgCh)
		msg, _ := sub.ReceiveMessage(ctx)
		msgCh <- msg
	}()

	err = pub.PublishJobUpdate(ctx, JobUpdateEvent{
		JobID:   "job-2",
		ImageID: imageID,
		Status:  "ready",
	})
	require.NoError(t, err)

	select {
	case msg := <-msgCh:
		require.NotNil(t, msg)
		assert.Equal(t, channel, msg.Channel)
		assert.JSONEq(t, `{"status":"ready"}`, msg.Payload)
	case <-ctx.Done():
		t.Fatal("timed out waiting for pubsub message")
	}
}
