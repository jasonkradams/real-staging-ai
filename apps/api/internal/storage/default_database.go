package storage

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/real-staging-ai/api/internal/config"
)

// DefaultDatabase wraps the pgx connection pool with tracing
type DefaultDatabase struct {
	pool   PgxPool
	tracer trace.Tracer
}

// NewDefaultDatabase creates a new database connection with OpenTelemetry instrumentation.
func NewDefaultDatabase(cfg *config.DB) (*DefaultDatabase, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is required")
	}

	ctx := context.Background()

	// Use URL if provided, otherwise construct from components
	var dbURL string
	if cfg.URL != "" {
		dbURL = cfg.URL
	} else {
		hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
		dbURL = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
			cfg.User, cfg.Password, hostPort, cfg.Database, cfg.SSLMode)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DefaultDatabase{
		pool:   pool,
		tracer: otel.Tracer("real-staging-api/database"),
	}, nil
}

// Close closes the database connection pool
func (db *DefaultDatabase) Close() {
	db.pool.Close()
}

// Pool returns the underlying connection pool for testing
func (db *DefaultDatabase) Pool() PgxPool {
	return db.pool
}

// QueryRow executes a query with tracing
func (db *DefaultDatabase) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	tr := db.tracer
	if tr == nil {
		tr = otel.Tracer("real-staging-api/database")
	}
	ctx, span := tr.Start(ctx, "db.query_row")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.statement", sql),
		attribute.Int("db.args.count", len(arguments)),
	)

	return db.pool.QueryRow(ctx, sql, arguments...)
}

// Query executes a query with tracing
func (db *DefaultDatabase) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	tr := db.tracer
	if tr == nil {
		tr = otel.Tracer("real-staging-api/database")
	}
	ctx, span := tr.Start(ctx, "db.query")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.statement", sql),
		attribute.Int("db.args.count", len(arguments)),
	)

	rows, err := db.pool.Query(ctx, sql, arguments...)
	if err != nil {
		span.RecordError(err)
	}

	return rows, err
}

// Exec executes a command with tracing
func (db *DefaultDatabase) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	tr := db.tracer
	if tr == nil {
		tr = otel.Tracer("real-staging-api/database")
	}
	ctx, span := tr.Start(ctx, "db.exec")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.statement", sql),
		attribute.Int("db.args.count", len(arguments)),
	)

	tag, err := db.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		span.RecordError(err)
	}

	return tag, err
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
