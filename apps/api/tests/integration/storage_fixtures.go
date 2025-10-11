//go:build integration

package integration

import (
	"context"
	_ "embed"

	"github.com/real-staging-ai/api/internal/storage"
)

//go:embed testdata/seed.sql
var seedSQL string

// SeedDatabase inserts test data into the database
func SeedDatabase(ctx context.Context, pool storage.PgxPool) error {
	_, err := pool.Exec(ctx, seedSQL)
	return err
}

// TruncateAllTables truncates all tables and resets sequences
func TruncateAllTables(ctx context.Context, pool storage.PgxPool) error {
	query := `
		TRUNCATE TABLE processed_events, invoices, subscriptions, images, jobs, projects, users, plans RESTART IDENTITY CASCADE
	`
	_, err := pool.Exec(ctx, query)
	return err
}

// ResetDatabase truncates all tables and seeds with test data
func ResetDatabase(ctx context.Context, pool storage.PgxPool) error {
	if err := TruncateAllTables(ctx, pool); err != nil {
		return err
	}
	return SeedDatabase(ctx, pool)
}
