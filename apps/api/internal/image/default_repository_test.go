package image

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

func TestDefaultRepository_CreateImage(t *testing.T) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return poolMock.QueryRow(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	projectID := uuid.New()
	roomType := "living_room"
	style := "modern"
	seed := int64(123)

	testCases := []struct {
		name        string
		projectID   string
		originalURL string
		roomType    *string
		style       *string
		seed        *int64
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:        "success: create image",
			projectID:   projectID.String(),
			originalURL: "http://example.com/image.jpg",
			roomType:    &roomType,
			style:       &style,
			seed:        &seed,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO images").
					WithArgs(
						pgtype.UUID{Bytes: projectID, Valid: true}, "http://example.com/image.jpg",
						pgtype.Text{String: "living_room", Valid: true},
						pgtype.Text{String: "modern", Valid: true},
						pgtype.Int8{Int64: 123, Valid: true},
					).
					WillReturnRows(
						pgxmock.NewRows([]string{
							"id", "project_id", "original_url", "staged_url",
							"room_type", "style", "seed", "status", "error", "created_at", "updated_at",
						}).
							AddRow(
								pgtype.UUID{Bytes: uuid.New(), Valid: true},
								pgtype.UUID{Bytes: projectID, Valid: true},
								"http://example.com/image.jpg", pgtype.Text{},
								pgtype.Text{String: "living_room", Valid: true},
								pgtype.Text{String: "modern", Valid: true},
								pgtype.Int8{Int64: 123, Valid: true},
								"queued", pgtype.Text{}, pgtype.Timestamptz{}, pgtype.Timestamptz{},
							))
			},
			expectError: false,
		},
		{
			name:        "fail: invalid project ID",
			projectID:   "invalid-uuid",
			originalURL: "http://example.com/image.jpg",
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:        "fail: query error",
			projectID:   projectID.String(),
			originalURL: "http://example.com/image.jpg",
			roomType:    &roomType,
			style:       &style,
			seed:        &seed,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO images").
					WithArgs(
						pgtype.UUID{Bytes: projectID, Valid: true}, "http://example.com/image.jpg",
						pgtype.Text{String: "living_room", Valid: true},
						pgtype.Text{String: "modern", Valid: true},
						pgtype.Int8{Int64: 123, Valid: true},
					).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.CreateImage(ctx, tc.projectID, tc.originalURL, tc.roomType, tc.style, tc.seed)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_GetImageByID(t *testing.T) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return poolMock.QueryRow(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	imageID := uuid.New()

	testCases := []struct {
		name        string
		imageID     string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:    "success: get image by id",
			imageID: imageID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(
					`-- name: GetImageByID :one\s+SELECT .+ FROM images\s+WHERE id = \$1`,
				).
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}).
					WillReturnRows(
						pgxmock.NewRows([]string{
							"id", "project_id", "original_url", "staged_url",
							"room_type", "style", "seed", "status", "error", "created_at", "updated_at",
						}).
							AddRow(
								pgtype.UUID{Bytes: uuid.New(), Valid: true},
								pgtype.UUID{Bytes: uuid.New(), Valid: true},
								"http://example.com/image.jpg", pgtype.Text{},
								pgtype.Text{String: "living_room", Valid: true},
								pgtype.Text{String: "modern", Valid: true},
								pgtype.Int8{Int64: 123, Valid: true},
								"queued", pgtype.Text{}, pgtype.Timestamptz{}, pgtype.Timestamptz{},
							))
			},
			expectError: false,
		},
		{
			name:        "fail: invalid image ID",
			imageID:     "invalid-uuid",
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:    "fail: query error",
			imageID: imageID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(
					`-- name: GetImageByID :one\s+SELECT .+ FROM images\s+WHERE id = \$1`,
				).
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
		{
			name:    "fail: not found",
			imageID: imageID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(
					`-- name: GetImageByID :one\s+SELECT .+ FROM images\s+WHERE id = \$1`,
				).
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}).
					WillReturnError(pgx.ErrNoRows)
			},
			expectError: true,
		},
		{
			name:    "fail: scan error",
			imageID: imageID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(
					`-- name: GetImageByID :one\s+SELECT .+ FROM images\s+WHERE id = \$1`,
				).
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("not a uuid"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.GetImageByID(ctx, tc.imageID)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_GetImagesByProjectID(t *testing.T) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		QueryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
			return poolMock.Query(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	projectID := uuid.New()

	testCases := []struct {
		name        string
		projectID   string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:      "success: get images by project id",
			projectID: projectID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(
					`-- name: GetImagesByProjectID :many\s+SELECT .+ FROM images\s+WHERE project_id = \$1\s+ORDER BY`,
				).
					WithArgs(pgtype.UUID{Bytes: projectID, Valid: true}).
					WillReturnRows(pgxmock.NewRows([]string{"id"}))
			},
			expectError: false,
		},
		{
			name:        "fail: invalid project ID",
			projectID:   "invalid-uuid",
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:      "fail: query error",
			projectID: projectID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(
					`-- name: GetImagesByProjectID :many\s+SELECT .+ FROM images\s+WHERE project_id = \$1\s+ORDER BY`,
				).
					WithArgs(pgtype.UUID{Bytes: projectID, Valid: true}).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.GetImagesByProjectID(ctx, tc.projectID)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_UpdateImageStatus(t *testing.T) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return poolMock.QueryRow(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	imageID := uuid.New()

	testCases := []struct {
		name        string
		imageID     string
		status      string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:    "success: update image status",
			imageID: imageID.String(),
			status:  "processing",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE images").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}, queries.ImageStatusProcessing).
					WillReturnRows(pgxmock.NewRows(
						[]string{"id", "project_id", "original_url", "staged_url", "room_type", "style",
							"seed", "status", "error", "created_at", "updated_at"}).
						AddRow(
							pgtype.UUID{Bytes: imageID, Valid: true},
							pgtype.UUID{Bytes: uuid.New(), Valid: true},
							"http://example.com/image.jpg",
							pgtype.Text{},
							pgtype.Text{String: "living_room", Valid: true},
							pgtype.Text{String: "modern", Valid: true},
							pgtype.Int8{Int64: 123, Valid: true},
							"processing",
							pgtype.Text{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{}))

			},
			expectError: false,
		},
		{
			name:        "fail: invalid image ID",
			imageID:     "invalid-uuid",
			status:      "processing",
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:    "fail: query error",
			imageID: imageID.String(),
			status:  "processing",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE images").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}, queries.ImageStatusProcessing).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.UpdateImageStatus(ctx, tc.imageID, tc.status)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_UpdateImageWithStagedURL(t *testing.T) {
	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbMock := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return mock.QueryRow(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	imageID := uuid.New()
	stagedURL := "http://example.com/staged.jpg"
	status := "ready"

	testCases := []struct {
		name        string
		imageID     string
		stagedURL   string
		status      string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:      "success: update image with staged url",
			imageID:   imageID.String(),
			stagedURL: stagedURL,
			status:    status,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE images").
					WithArgs(
						pgtype.UUID{Bytes: imageID, Valid: true},
						pgtype.Text{String: stagedURL, Valid: true},
						queries.ImageStatusReady).
					WillReturnRows(pgxmock.NewRows(
						[]string{"id", "project_id", "original_url", "staged_url", "room_type", "style",
							"seed", "status", "error", "created_at", "updated_at"}).
						AddRow(
							pgtype.UUID{Bytes: imageID, Valid: true},
							pgtype.UUID{Bytes: uuid.New(), Valid: true},
							"http://example.com/image.jpg",
							pgtype.Text{String: stagedURL, Valid: true},
							pgtype.Text{String: "living_room", Valid: true},
							pgtype.Text{String: "modern", Valid: true},
							pgtype.Int8{Int64: 123, Valid: true},
							"ready",
							pgtype.Text{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{}))

			},
			expectError: false,
		},
		{
			name:        "fail: invalid image ID",
			imageID:     "invalid-uuid",
			stagedURL:   stagedURL,
			status:      status,
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:      "fail: query error",
			imageID:   imageID.String(),
			stagedURL: stagedURL,
			status:    status,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE images").
					WithArgs(
						pgtype.UUID{Bytes: imageID, Valid: true},
						pgtype.Text{String: stagedURL, Valid: true},
						queries.ImageStatusReady).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(mock)
			_, err := repo.UpdateImageWithStagedURL(ctx, tc.imageID, tc.stagedURL, tc.status)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_UpdateImageWithError(t *testing.T) {
	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbMock := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return mock.QueryRow(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	imageID := uuid.New()
	errorMsg := "something went wrong"

	testCases := []struct {
		name        string
		imageID     string
		errorMsg    string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:     "success: update image with error",
			imageID:  imageID.String(),
			errorMsg: errorMsg,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE images").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}, pgtype.Text{String: errorMsg, Valid: true}).
					WillReturnRows(pgxmock.NewRows(
						[]string{"id",
							"project_id",
							"original_url",
							"staged_url",
							"room_type",
							"style",
							"seed",
							"status",
							"error",
							"created_at",
							"updated_at"}).
						AddRow(
							pgtype.UUID{Bytes: imageID, Valid: true},
							pgtype.UUID{Bytes: uuid.New(), Valid: true},
							"http://example.com/image.jpg",
							pgtype.Text{},
							pgtype.Text{String: "living_room", Valid: true},
							pgtype.Text{String: "modern", Valid: true},
							pgtype.Int8{Int64: 123, Valid: true},
							"error",
							pgtype.Text{String: errorMsg, Valid: true},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{}))
			},
			expectError: false,
		},
		{
			name:        "fail: invalid image ID",
			imageID:     "invalid-uuid",
			errorMsg:    errorMsg,
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:     "fail: query error",
			imageID:  imageID.String(),
			errorMsg: errorMsg,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE images").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}, pgtype.Text{String: errorMsg, Valid: true}).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(mock)
			_, err := repo.UpdateImageWithError(ctx, tc.imageID, tc.errorMsg)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_DeleteImage(t *testing.T) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		ExecFunc: func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
			return poolMock.Exec(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	imageID := uuid.New()

	testCases := []struct {
		name        string
		imageID     string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:    "success: delete image",
			imageID: imageID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM images").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectError: false,
		},
		{
			name:        "fail: invalid image ID",
			imageID:     "invalid-uuid",
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:    "fail: query error",
			imageID: imageID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM images").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
		{
			name:    "fail: image not found",
			imageID: imageID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM images").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectError: false, // DeleteImage doesn't return error for 0 rows affected
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			err := repo.DeleteImage(ctx, tc.imageID)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_DeleteImagesByProjectID(t *testing.T) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		ExecFunc: func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
			return poolMock.Exec(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	projectID := uuid.New()

	testCases := []struct {
		name        string
		projectID   string
		setupMock   func()
		expectError bool
		errorMsg    string
	}{
		{
			name:      "success: delete images by project id",
			projectID: projectID.String(),
			setupMock: func() {
				poolMock.ExpectExec("DELETE FROM images WHERE project_id = \\$1").
					WithArgs(pgtype.UUID{Bytes: projectID, Valid: true}).
					WillReturnResult(pgxmock.NewResult("DELETE", 2))
			},
			expectError: false,
		},
		{
			name:        "fail: invalid project ID",
			projectID:   "invalid-uuid",
			setupMock:   func() {},
			expectError: true,
			errorMsg:    "invalid project ID",
		},
		{
			name:      "fail: query error",
			projectID: projectID.String(),
			setupMock: func() {
				poolMock.ExpectExec("DELETE FROM images WHERE project_id = \\$1").
					WithArgs(pgtype.UUID{Bytes: projectID, Valid: true}).
					WillReturnError(errors.New("database error"))
			},
			expectError: true,
			errorMsg:    "failed to delete images",
		},
		{
			name:      "success: no images found (0 rows affected)",
			projectID: projectID.String(),
			setupMock: func() {
				poolMock.ExpectExec("DELETE FROM images WHERE project_id = \\$1").
					WithArgs(pgtype.UUID{Bytes: projectID, Valid: true}).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			err := repo.DeleteImagesByProjectID(ctx, tc.projectID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}
