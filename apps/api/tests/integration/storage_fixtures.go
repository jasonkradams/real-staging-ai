//go:build integration

package integration

import (
	"context"
	_ "embed"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// This path is relative to the internal/storage directory
var seedSQL, _ = os.ReadFile("../../testdata/seed.sql")

// SeedDatabase inserts test data into the database
func SeedDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, string(seedSQL))
	return err
}

// TruncateAllTables truncates all tables and resets sequences
func TruncateAllTables(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		TRUNCATE TABLE projects, users, images, jobs, plans RESTART IDENTITY
	`
	_, err := pool.Exec(ctx, query)
	return err
}

// ResetDatabase truncates all tables and seeds with test data
func ResetDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	if err := TruncateAllTables(ctx, pool); err != nil {
		return err
	}
	return SeedDatabase(ctx, pool)
}
