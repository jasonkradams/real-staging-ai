package telemetry

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitTracing(t *testing.T) {
	tests := []struct {
		name          string
		setup         func()
		getCtx        func() (context.Context, context.CancelFunc)
		expectedErr   bool
		expectedPanic bool
	}{
		{
			name: "success: with default endpoint",
			setup: func() {
				err := os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
				if err != nil {
					panic(err)
				}
			},
			getCtx: func() (context.Context, context.CancelFunc) {
				return context.Background(), func() {}
			},
			expectedErr:   false,
			expectedPanic: false,
		},
		{
			name: "success: with custom endpoint",
			setup: func() {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4319")
			},
			getCtx: func() (context.Context, context.CancelFunc) {
				return context.Background(), func() {}
			},
			expectedErr:   false,
			expectedPanic: false,
		},
		{
			name:  "fail: resource creation fails with canceled context",
			setup: func() {},
			getCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, func() {}
			},
			expectedErr:   true,
			expectedPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			ctx, cancel := tt.getCtx()
			defer cancel()

			if tt.expectedPanic {
				assert.Panics(t, func() {
					_, _ = InitTracing(ctx, "test-service")
				})
				return
			}

			shutdown, err := InitTracing(ctx, "test-service")

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, shutdown)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, shutdown)
				assert.NoError(t, shutdown(context.Background()))
			}
		})
	}
}
