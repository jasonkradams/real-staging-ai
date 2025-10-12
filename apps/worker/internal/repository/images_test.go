package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (*DefaultImageRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewImageRepository(db)
	cleanup := func() {
		_ = db.Close()
	}
	return repo, mock, cleanup
}

func TestDefaultImageRepository_SetProcessing_Success(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ctx := context.Background()
	imageID := "2e1aa86e-0f27-4f0f-9e57-0d7b53b4d9b9"

	// Match multi-line SQL and allow whitespace/newlines
	query := regexp.QuoteMeta(
		"UPDATE images SET status = 'processing', updated_at = now() " +
			"WHERE id = $1::uuid AND status IN ('queued','processing');")
	mock.ExpectExec(query).
		WithArgs(imageID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetProcessing(ctx, imageID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultImageRepository_SetProcessing_DBError(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ctx := context.Background()
	imageID := "2e1aa86e-0f27-4f0f-9e57-0d7b53b4d9b9"

	query := regexp.QuoteMeta(
		"UPDATE images SET status = 'processing', updated_at = now() " +
			"WHERE id = $1::uuid AND status IN ('queued','processing');")
	mock.ExpectExec(query).
		WithArgs(imageID).
		WillReturnError(assert.AnError)

	err := repo.SetProcessing(ctx, imageID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update image status to processing")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultImageRepository_SetReady_Success(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ctx := context.Background()
	imageID := "b5a4b7a1-3584-4b09-9b6a-6a6e1f2d9e90"
	stagedURL := "https://example.com/image-staged.jpg"

	query := regexp.QuoteMeta(
		"UPDATE images SET staged_url = $2, status = 'ready', updated_at = now() " +
			"WHERE id = $1::uuid AND status IN ('queued','processing');")
	mock.ExpectExec(query).
		WithArgs(imageID, stagedURL).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetReady(ctx, imageID, stagedURL)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultImageRepository_SetReady_EmptyURL(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ctx := context.Background()
	imageID := "b5a4b7a1-3584-4b09-9b6a-6a6e1f2d9e90"

	err := repo.SetReady(ctx, imageID, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stagedURL cannot be empty")
	// No SQL should have been executed
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultImageRepository_SetReady_DBError(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ctx := context.Background()
	imageID := "b5a4b7a1-3584-4b09-9b6a-6a6e1f2d9e90"
	stagedURL := "https://example.com/image-staged.jpg"

	query := regexp.QuoteMeta(
		"UPDATE images SET staged_url = $2, status = 'ready', updated_at = now() " +
			"WHERE id = $1::uuid AND status IN ('queued','processing');")
	mock.ExpectExec(query).
		WithArgs(imageID, stagedURL).
		WillReturnError(assert.AnError)

	err := repo.SetReady(ctx, imageID, stagedURL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update image with staged url")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultImageRepository_SetError_Success(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ctx := context.Background()
	imageID := "0e5b2e97-4324-4f47-bc8b-05d33d62d9b4"
	errMsg := "processing failed"

	query := regexp.QuoteMeta(
		"UPDATE images SET status = 'error', error = $2, updated_at = now() " +
			"WHERE id = $1::uuid AND status IN ('queued','processing');")
	mock.ExpectExec(query).
		WithArgs(imageID, errMsg).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetError(ctx, imageID, errMsg)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultImageRepository_SetError_EmptyMsg(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ctx := context.Background()
	imageID := "0e5b2e97-4324-4f47-bc8b-05d33d62d9b4"

	err := repo.SetError(ctx, imageID, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error message cannot be empty")
	// No SQL should have been executed
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultImageRepository_SetError_DBError(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ctx := context.Background()
	imageID := "0e5b2e97-4324-4f47-bc8b-05d33d62d9b4"
	errMsg := "processing failed"

	query := regexp.QuoteMeta(
		"UPDATE images SET status = 'error', error = $2, updated_at = now() " +
			"WHERE id = $1::uuid AND status IN ('queued','processing');")
	mock.ExpectExec(query).
		WithArgs(imageID, errMsg).
		WillReturnError(assert.AnError)

	err := repo.SetError(ctx, imageID, errMsg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update image with error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
