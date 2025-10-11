//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/real-staging-ai/api/internal/reconcile"
)

func TestReconcileImages_Integration(t *testing.T) {
	ctx := context.Background()

	// Setup database
	db := SetupTestDatabase(t)
	defer db.Close()

	TruncateAllTables(ctx, db.Pool())
	SeedDatabase(ctx, db.Pool())

	// Setup S3 service (LocalStack)
	s3Service := SetupTestS3Service(t, ctx)

	// Create service
	svc := reconcile.NewDefaultService(db, s3Service)

	t.Run("success: detects missing original file", func(t *testing.T) {
		// Create test user and project
		userID := uuid.New()
		projectID := uuid.New()
		imageID := uuid.New()

		_, _ = db.Pool().Exec(ctx, `INSERT INTO users (id, auth0_sub) VALUES ($1, $2)`,
			userID, "auth0|test-reconcile-1")

		_, _ = db.Pool().Exec(ctx, `INSERT INTO projects (id, user_id, name) VALUES ($1, $2, $3)`,
			projectID, userID, "Test Project Reconcile")

		// Insert image with original_url that doesn't exist in S3
		originalURL := fmt.Sprintf("http://localhost:4566/real-staging/uploads/%s/original.jpg", imageID)
		_, _ = db.Pool().Exec(ctx, `
			INSERT INTO images (id, project_id, original_url, status, created_at, updated_at)
			VALUES ($1, $2, $3, 'queued', NOW(), NOW())`,
			imageID, projectID, originalURL)

		// Run reconciliation
		result, err := svc.ReconcileImages(ctx, reconcile.ReconcileOptions{
			DryRun:      false,
			Concurrency: 2,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Checked, 1, "should check at least 1 image")
		assert.GreaterOrEqual(t, result.MissingOrig, 1, "should detect at least 1 missing original")
		assert.GreaterOrEqual(t, result.Updated, 1, "should update at least 1 image")

		// Verify database was updated
		var status, errorMsg string
		err = db.Pool().QueryRow(ctx, `SELECT status, error FROM images WHERE id = $1`, imageID).
			Scan(&status, &errorMsg)
		require.NoError(t, err)
		assert.Equal(t, "error", status)
		assert.Contains(t, errorMsg, "original missing in storage")

		// Cleanup
		_, _ = db.Pool().Exec(ctx, `DELETE FROM images WHERE id = $1`, imageID)
		_, _ = db.Pool().Exec(ctx, `DELETE FROM projects WHERE id = $1`, projectID)
		_, _ = db.Pool().Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})

	t.Run("success: dry run mode does not update database", func(t *testing.T) {
		// Create test user and project
		userID := uuid.New()
		projectID := uuid.New()
		imageID := uuid.New()

		_, _ = db.Pool().Exec(ctx, `INSERT INTO users (id, auth0_sub) VALUES ($1, $2)`,
			userID, "auth0|test-reconcile-2")

		_, _ = db.Pool().Exec(ctx, `INSERT INTO projects (id, user_id, name) VALUES ($1, $2, $3)`,
			projectID, userID, "Test Project Reconcile 2")

		// Insert image with original_url that doesn't exist
		originalURL := fmt.Sprintf("http://localhost:4566/real-staging/uploads/%s/original.jpg", imageID)
		_, _ = db.Pool().Exec(ctx, `
			INSERT INTO images (id, project_id, original_url, status, created_at, updated_at)
			VALUES ($1, $2, $3, 'queued', NOW(), NOW())`,
			imageID, projectID, originalURL)

		// Run reconciliation in dry-run mode
		result, err := svc.ReconcileImages(ctx, reconcile.ReconcileOptions{
			DryRun:      true,
			Concurrency: 2,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Checked, 1)
		assert.GreaterOrEqual(t, result.MissingOrig, 1)
		assert.Equal(t, 0, result.Updated, "dry run should not update")
		assert.True(t, result.DryRun)

		// Verify database was NOT updated
		var status string
		var errorMsg *string
		err = db.Pool().QueryRow(ctx, `SELECT status, error FROM images WHERE id = $1`, imageID).
			Scan(&status, &errorMsg)
		require.NoError(t, err)
		assert.Equal(t, "queued", status, "status should remain unchanged")
		assert.Nil(t, errorMsg, "error should be nil")

		// Cleanup
		_, _ = db.Pool().Exec(ctx, `DELETE FROM images WHERE id = $1`, imageID)
		_, _ = db.Pool().Exec(ctx, `DELETE FROM projects WHERE id = $1`, projectID)
		_, _ = db.Pool().Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})

	t.Run("success: filters by project_id", func(t *testing.T) {
		// Create test users and projects
		userID := uuid.New()
		project1 := uuid.New()
		project2 := uuid.New()
		image1 := uuid.New()
		image2 := uuid.New()

		_, _ = db.Pool().Exec(ctx, `INSERT INTO users (id, auth0_sub) VALUES ($1, $2)`,
			userID, "auth0|test-reconcile-3")

		_, _ = db.Pool().Exec(ctx, `INSERT INTO projects (id, user_id, name) VALUES ($1, $2, $3), ($4, $2, $5)`,
			project1, userID, "Project 1", project2, "Project 2")

		// Insert images for both projects
		originalURL1 := fmt.Sprintf("http://localhost:4566/real-staging/uploads/%s/original.jpg", image1)
		originalURL2 := fmt.Sprintf("http://localhost:4566/real-staging/uploads/%s/original.jpg", image2)

		_, _ = db.Pool().Exec(ctx, `
			INSERT INTO images (id, project_id, original_url, status, created_at, updated_at)
			VALUES ($1, $2, $3, 'queued', NOW(), NOW()), ($4, $5, $6, 'queued', NOW(), NOW())`,
			image1, project1, originalURL1, image2, project2, originalURL2)

		// Run reconciliation filtered by project1
		projectIDStr := project1.String()
		result, err := svc.ReconcileImages(ctx, reconcile.ReconcileOptions{
			ProjectID:   &projectIDStr,
			DryRun:      true,
			Concurrency: 2,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Checked, "should only check images from project1")

		// Cleanup
		_, _ = db.Pool().Exec(ctx, `DELETE FROM images WHERE id IN ($1, $2)`, image1, image2)
		_, _ = db.Pool().Exec(ctx, `DELETE FROM projects WHERE id IN ($1, $2)`, project1, project2)
		_, _ = db.Pool().Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})
}
