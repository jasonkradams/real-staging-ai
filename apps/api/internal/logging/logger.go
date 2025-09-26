package logging

import "context"

//go:generate go run github.com/matryer/moq@v0.5.3 -out logger_mock.go . Logger

// Logger is a minimal structured logging interface for the API.
// Implementations should include OTEL trace correlation (trace_id, span_id)
// when a context carries a Span.
type Logger interface {
	Info(ctx context.Context, msg string, keysAndValues ...any)
	Warn(ctx context.Context, msg string, keysAndValues ...any)
	Error(ctx context.Context, msg string, keysAndValues ...any)
	Debug(ctx context.Context, msg string, keysAndValues ...any)
}

var defaultLogger Logger = NewDefaultLogger()

// Default returns the process-wide default Logger.
func Default() Logger { return defaultLogger }

// SetDefault overrides the process-wide default Logger (useful for tests).
func SetDefault(l Logger) {
	if l != nil {
		defaultLogger = l
	}
}
