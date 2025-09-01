package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TruncateTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	query := `
		TRUNCATE TABLE projects, users, images, jobs, plans RESTART IDENTITY
	`
	_, err := pool.Exec(context.Background(), query)
	require.NoError(t, err)
}

func SeedTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// This path is relative to the package that is being tested.
	// This is not ideal, but it's the simplest solution for now.
	seedSQL, err := os.ReadFile("../../testdata/seed.sql")
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), string(seedSQL))
	require.NoError(t, err)
}
