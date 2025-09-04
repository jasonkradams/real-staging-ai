//go:build integration

package integration

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed testdata/seed.sql
var seedSQL string

// SeedDatabase inserts test data into the database
func SeedDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, seedSQL)
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
