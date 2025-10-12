package job

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to test QueryRow-based job operations (GetJobByID, StartJob, CompleteJob)
func testJobQueryRowOperation(
	t *testing.T,
	operationName string,
	queryName string,
	expectedStatus string,
	operation func(repo *DefaultRepository, ctx context.Context, jobID string) (*queries.Job, error),
) {
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
			name:  "success: " + operationName,
			jobID: jobID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(queryName).
					WithArgs(pgtype.UUID{Bytes: jobID, Valid: true}).
					WillReturnRows(pgxmock.NewRows(
						[]string{"id", "image_id", "type", "payload_json", "status",
							"error", "created_at", "started_at", "finished_at"}).
						AddRow(
							pgtype.UUID{Bytes: jobID, Valid: true},
							pgtype.UUID{Bytes: imageID, Valid: true},
							jobType,
							payloadJSON,
							expectedStatus,
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
				mock.ExpectQuery(queryName).
					WithArgs(pgtype.UUID{Bytes: jobID, Valid: true}).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := operation(repo, ctx, tc.jobID)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

// Helper to test Exec-based job operations (DeleteJob, DeleteJobsByImageID)
func testJobExecOperation(
	t *testing.T,
	operationName string,
	queryName string,
	idParamName string,
	operation func(repo *DefaultRepository, ctx context.Context, id string) error,
) {
	ctx := context.Background()
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	dbMock := &storage.DatabaseMock{
		ExecFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			return poolMock.Exec(ctx, sql, args...)
		},
	}

	repo := NewDefaultRepository(dbMock)
	testID := uuid.New()

	testCases := []struct {
		name        string
		id          string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name: "success: " + operationName,
			id:   testID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(queryName).
					WithArgs(pgtype.UUID{Bytes: testID, Valid: true}).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectError: false,
		},
		{
			name:        "fail: invalid " + idParamName,
			id:          "invalid-uuid",
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name: "fail: query error",
			id:   testID.String(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(queryName).
					WithArgs(pgtype.UUID{Bytes: testID, Valid: true}).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			err := operation(repo, ctx, tc.id)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

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
					WillReturnRows(pgxmock.NewRows(
						[]string{"id", "image_id", "type", "payload_json", "status",
							"error", "created_at", "started_at", "finished_at"}).
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
				mock.ExpectQuery("INSERT INTO jobs").
					WithArgs(pgtype.UUID{Bytes: imageID, Valid: true}, jobType, payloadJSON).
					WillReturnError(errors.New("db error"))
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
	testJobQueryRowOperation(t, "get job by id", "GetJobByID", "queued",
		func(repo *DefaultRepository, ctx context.Context, jobID string) (*queries.Job, error) {
			return repo.GetJobByID(ctx, jobID)
		})
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
					WillReturnRows(pgxmock.NewRows(
						[]string{"id", "image_id", "type", "payload_json", "status",
							"error", "created_at", "started_at", "finished_at"}).
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

func TestDefaultRepository_UpdateJobStatus(t *testing.T) {
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
	status := "processing"

	testCases := []struct {
		name        string
		jobID       string
		status      string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:   "success: update job status",
			jobID:  jobID.String(),
			status: status,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UpdateJobStatus").
					WithArgs(pgtype.UUID{Bytes: jobID, Valid: true}, status).
					WillReturnRows(pgxmock.NewRows(
						[]string{"id", "image_id", "type", "payload_json", "status",
							"error", "created_at", "started_at", "finished_at"}).
						AddRow(
							pgtype.UUID{Bytes: jobID, Valid: true},
							pgtype.UUID{Bytes: imageID, Valid: true},
							jobType,
							payloadJSON,
							status,
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
			status:      status,
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:   "fail: query error",
			jobID:  jobID.String(),
			status: status,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UpdateJobStatus").
					WithArgs(pgtype.UUID{Bytes: jobID, Valid: true}, status).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.UpdateJobStatus(ctx, tc.jobID, tc.status)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_StartJob(t *testing.T) {
	testJobQueryRowOperation(t, "start job", "StartJob", "processing",
		func(repo *DefaultRepository, ctx context.Context, jobID string) (*queries.Job, error) {
			return repo.StartJob(ctx, jobID)
		})
}

func TestDefaultRepository_CompleteJob(t *testing.T) {
	testJobQueryRowOperation(t, "complete job", "CompleteJob", "completed",
		func(repo *DefaultRepository, ctx context.Context, jobID string) (*queries.Job, error) {
			return repo.CompleteJob(ctx, jobID)
		})
}

func TestDefaultRepository_FailJob(t *testing.T) {
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
	errorMsg := "job failed"

	testCases := []struct {
		name        string
		jobID       string
		errorMsg    string
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:     "success: fail job",
			jobID:    jobID.String(),
			errorMsg: errorMsg,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("FailJob").
					WithArgs(pgtype.UUID{Bytes: jobID, Valid: true}, pgtype.Text{String: errorMsg, Valid: true}).
					WillReturnRows(pgxmock.NewRows(
						[]string{"id", "image_id", "type", "payload_json", "status",
							"error", "created_at", "started_at", "finished_at"}).
						AddRow(
							pgtype.UUID{Bytes: jobID, Valid: true},
							pgtype.UUID{Bytes: imageID, Valid: true},
							jobType,
							payloadJSON,
							"failed",
							pgtype.Text{String: errorMsg, Valid: true},
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
			errorMsg:    errorMsg,
			setupMock:   func(mock pgxmock.PgxPoolIface) {},
			expectError: true,
		},
		{
			name:     "fail: query error",
			jobID:    jobID.String(),
			errorMsg: errorMsg,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("FailJob").
					WithArgs(pgtype.UUID{Bytes: jobID, Valid: true}, pgtype.Text{String: errorMsg, Valid: true}).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.FailJob(ctx, tc.jobID, tc.errorMsg)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_GetPendingJobs(t *testing.T) {
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
	limit := 10

	testCases := []struct {
		name        string
		limit       int
		setupMock   func(mock pgxmock.PgxPoolIface)
		expectError bool
	}{
		{
			name:  "success: get pending jobs",
			limit: limit,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("GetPendingJobs").
					WithArgs(int32(limit)).
					WillReturnRows(pgxmock.NewRows(
						[]string{"id", "image_id", "type", "payload_json", "status",
							"error", "created_at", "started_at", "finished_at"}).
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
			name:  "fail: query error",
			limit: limit,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("GetPendingJobs").
					WithArgs(int32(limit)).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(poolMock)
			_, err := repo.GetPendingJobs(ctx, tc.limit)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, poolMock.ExpectationsWereMet())
		})
	}
}

func TestDefaultRepository_DeleteJob(t *testing.T) {
	testJobExecOperation(t, "delete job", "DeleteJob", "job ID",
		func(repo *DefaultRepository, ctx context.Context, id string) error {
			return repo.DeleteJob(ctx, id)
		})
}

func TestDefaultRepository_DeleteJobsByImageID(t *testing.T) {
	testJobExecOperation(t, "delete jobs by image id", "DeleteJobsByImageID", "image ID",
		func(repo *DefaultRepository, ctx context.Context, id string) error {
			return repo.DeleteJobsByImageID(ctx, id)
		})
}
