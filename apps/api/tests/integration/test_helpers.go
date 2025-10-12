//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/stretchr/testify/require"
)

// SetupTestDatabase creates a database connection for integration tests using config.
func SetupTestDatabase(t *testing.T) *storage.DefaultDatabase {
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err, "failed to load config")

	db, err := storage.NewDefaultDatabase(&cfg.DB)
	require.NoError(t, err, "failed to create database connection")

	return db
}

// SetupTestS3Service creates an S3 service for integration tests using config.
func SetupTestS3Service(t *testing.T, ctx context.Context) *storage.DefaultS3Service {
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err, "failed to load config")

	s3Service, err := storage.NewDefaultS3Service(ctx, &cfg.S3)
	require.NoError(t, err, "failed to create S3 service")

	return s3Service
}
