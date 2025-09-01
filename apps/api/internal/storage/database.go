package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Database wraps the pgx connection pool with tracing
type Database struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

// NewDatabase creates a new database connection with OpenTelemetry instrumentation
func NewDatabase() (*Database, error) {
	ctx := context.Background()

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Construct from individual components
		host := getEnvOrDefault("PGHOST", "localhost")
		port := getEnvOrDefault("PGPORT", "5432")
		user := getEnvOrDefault("PGUSER", "postgres")
		password := getEnvOrDefault("PGPASSWORD", "postgres")
		dbname := getEnvOrDefault("PGDATABASE", "virtualstaging")
		sslmode := getEnvOrDefault("PGSSLMODE", "disable")

		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbname, sslmode)
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

	return &Database{
		pool:   pool,
		tracer: otel.Tracer("virtual-staging-api/database"),
	}, nil
}

// Close closes the database connection pool
func (db *Database) Close() {
	db.pool.Close()
}

// QueryRow executes a query with tracing
func (db *Database) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	ctx, span := db.tracer.Start(ctx, "db.query_row")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.statement", sql),
		attribute.Int("db.args.count", len(args)),
	)

	return db.pool.QueryRow(ctx, sql, args...)
}

// Query executes a query with tracing
func (db *Database) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	ctx, span := db.tracer.Start(ctx, "db.query")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.statement", sql),
		attribute.Int("db.args.count", len(args)),
	)

	rows, err := db.pool.Query(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
	}

	return rows, err
}

// Exec executes a command with tracing
func (db *Database) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	ctx, span := db.tracer.Start(ctx, "db.exec")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.statement", sql),
		attribute.Int("db.args.count", len(args)),
	)

	tag, err := db.pool.Exec(ctx, sql, args...)
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
