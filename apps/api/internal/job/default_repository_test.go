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
					WillReturnRows(pgxmock.NewRows([]string{"id", "image_id", "type", "payload_json", "status", "created_at", "updated_at"}).
						AddRow(pgtype.UUID{Bytes: uuid.New(), Valid: true}, pgtype.UUID{Bytes: imageID, Valid: true}, jobType, payloadJSON, "queued", pgtype.Timestamptz{}, pgtype.Timestamptz{}))
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
