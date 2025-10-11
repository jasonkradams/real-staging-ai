package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

type defaultSLogger struct {
	l *slog.Logger
}

// NewDefaultLogger constructs a slog-backed Logger with JSON output to stdout.
func NewDefaultLogger() Logger {
	lvl := parseLevel(os.Getenv("LOG_LEVEL"))
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	l := slog.New(h).With("service", "real-staging-api")
	return &defaultSLogger{l: l}
}

func (d *defaultSLogger) withTrace(ctx context.Context) *slog.Logger {
	l := d.l
	sc := trace.SpanContextFromContext(ctx)
	if sc.HasTraceID() {
		l = l.With("trace_id", sc.TraceID().String())
	}
	if sc.HasSpanID() {
		l = l.With("span_id", sc.SpanID().String())
	}
	return l
}

func (d *defaultSLogger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	d.withTrace(ctx).Info(msg, keysAndValues...)
}

func (d *defaultSLogger) Warn(ctx context.Context, msg string, keysAndValues ...any) {
	d.withTrace(ctx).Warn(msg, keysAndValues...)
}

func (d *defaultSLogger) Error(ctx context.Context, msg string, keysAndValues ...any) {
	d.withTrace(ctx).Error(msg, keysAndValues...)
}

func (d *defaultSLogger) Debug(ctx context.Context, msg string, keysAndValues ...any) {
	d.withTrace(ctx).Debug(msg, keysAndValues...)
}

// parseLevel converts LOG_LEVEL into a slog.Leveler
func parseLevel(s string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		lvl := slog.LevelDebug
		return &lvl
	case "warn", "warning":
		lvl := slog.LevelWarn
		return &lvl
	case "error":
		lvl := slog.LevelError
		return &lvl
	case "info", "":
		fallthrough
	default:
		lvl := slog.LevelInfo
		return &lvl
	}
}
