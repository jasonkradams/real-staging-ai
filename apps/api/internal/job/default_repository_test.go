package job

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-staging-ai/api/internal/storage"
)

func TestDefaultRepository_CreateJob(t *testing.T) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return poolMock.QueryRow(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	imageID := uuid.New()
	jobType := "stage:run"
	payload := map[string]any{"key": "value"}
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		imageID     string
		jobType     string
		payloadJSON []byte
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:        "success: create job",
			imageID:     imageID.String(),
			jobType:     jobType,
			payloadJSON: payloadJSON,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO jobs").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}, jobType, payloadJSON).
					WillReturnRows(pgxmock.NewRows([]string{"id", "image_id", "type", "payload_json", "status", "error", "created_at", "started_at", "finished_at"}).
						AddRow(
							pgtype.UUID{Bytes: uuid.New(), Valid: true},
							pgtype.UUID{Bytes: imageID, Valid: true},
							jobType,
							payloadJSON,
							"queued",
							pgtype.Text{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{},
						))
			},
			expectError: false,
		},
		{
			name:        "fail: invalid image ID",
			imageID:     "invalid-uuid",
			jobType:     jobType,
			payloadJSON: payloadJSON,
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:        "fail: query error",
			imageID:     imageID.String(),
			jobType:     jobType,
			payloadJSON: payloadJSON,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO jobs").WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}, jobType, payloadJSON).WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.CreateJob(ctx, tc.imageID, tc.jobType, tc.payloadJSON)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_GetJobByID(t *testing.T) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return poolMock.QueryRow(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	jobID := uuid.New()
	imageID := uuid.New()
	jobType := "stage:run"
	payload := map[string]any{"key": "value"}
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		jobID       string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:  "success: get job by id",
			jobID: jobID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("GetJobByID").
					WithArgs(pgtype.UUID{Bytes: jobID, Valid: true}).
					WillReturnRows(pgxmock.NewRows([]string{"id", "image_id", "type", "payload_json", "status", "error", "created_at", "started_at", "finished_at"}).
						AddRow(
							pgtype.UUID{Bytes: jobID, Valid: true},
							pgtype.UUID{Bytes: imageID, Valid: true},
							jobType,
							payloadJSON,
							"queued",
							pgtype.Text{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{},
						))
			},
			expectError: false,
		},
		{
			name:        "fail: invalid job ID",
			jobID:       "invalid-uuid",
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:  "fail: query error",
			jobID: jobID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("GetJobByID").
					WithArgs(pgtype.UUID{Bytes: jobID, Valid: true}).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.GetJobByID(ctx, tc.jobID)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_GetJobsByImageID(t *testing.T) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		QueryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			return poolMock.Query(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)

	jobID1 := uuid.New()
	jobID2 := uuid.New()
	imageID := uuid.New()
	jobType := "stage:run"
	payload := map[string]any{"key": "value"}
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		imageID     string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:    "success: get jobs by image id",
			imageID: imageID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("GetJobsByImageID").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}).
					WillReturnRows(pgxmock.NewRows([]string{"id", "image_id", "type", "payload_json", "status", "error", "created_at", "started_at", "finished_at"}).
						AddRow(
							pgtype.UUID{Bytes: jobID1, Valid: true},
							pgtype.UUID{Bytes: imageID, Valid: true},
							jobType,
							payloadJSON,
							"queued",
							pgtype.Text{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{},
						).
						AddRow(
							pgtype.UUID{Bytes: jobID2, Valid: true},
							pgtype.UUID{Bytes: imageID, Valid: true},
							jobType,
							payloadJSON,
							"completed",
							pgtype.Text{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{},
							pgtype.Timestamptz{},
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
				mock.ExpectQuery("GetJobsByImageID").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.GetJobsByImageID(ctx, tc.imageID)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}
