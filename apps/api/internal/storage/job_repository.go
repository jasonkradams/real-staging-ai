package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// JobRepositoryImpl implements the JobRepository interface.
type JobRepositoryImpl struct {
	db *DB
}

// Ensure JobRepositoryImpl implements JobRepository interface.
var _ JobRepository = (*JobRepositoryImpl)(nil)

// NewJobRepository creates a new JobRepositoryImpl instance.
func NewJobRepository(db *DB) *JobRepositoryImpl {
	return &JobRepositoryImpl{db: db}
}

// CreateJob creates a new job in the database.
func (r *JobRepositoryImpl) CreateJob(ctx context.Context, imageID string, jobType string, payloadJSON []byte) (*queries.Job, error) {
	q := queries.New(r.db.pool)

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %w", err)
	}

	job, err := q.CreateJob(ctx, queries.CreateJobParams{
		ImageID:     pgtype.UUID{Bytes: imageUUID, Valid: true},
		Type:        jobType,
		PayloadJson: payloadJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return job, nil
}

// GetJobByID retrieves a specific job by its ID.
func (r *JobRepositoryImpl) GetJobByID(ctx context.Context, jobID string) (*queries.Job, error) {
	q := queries.New(r.db.pool)

	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	job, err := q.GetJobByID(ctx, pgtype.UUID{Bytes: jobUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

// GetJobsByImageID retrieves all jobs for a specific image.
func (r *JobRepositoryImpl) GetJobsByImageID(ctx context.Context, imageID string) ([]*queries.Job, error) {
	q := queries.New(r.db.pool)

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %w", err)
	}

	jobs, err := q.GetJobsByImageID(ctx, pgtype.UUID{Bytes: imageUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	return jobs, nil
}

// UpdateJobStatus updates a job's status.
func (r *JobRepositoryImpl) UpdateJobStatus(ctx context.Context, jobID string, status string) (*queries.Job, error) {
	q := queries.New(r.db.pool)

	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	job, err := q.UpdateJobStatus(ctx, queries.UpdateJobStatusParams{
		ID:     pgtype.UUID{Bytes: jobUUID, Valid: true},
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update job status: %w", err)
	}

	return job, nil
}

// StartJob marks a job as processing and sets the started timestamp.
func (r *JobRepositoryImpl) StartJob(ctx context.Context, jobID string) (*queries.Job, error) {
	q := queries.New(r.db.pool)

	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	job, err := q.StartJob(ctx, pgtype.UUID{Bytes: jobUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to start job: %w", err)
	}

	return job, nil
}

// CompleteJob marks a job as completed and sets the finished timestamp.
func (r *JobRepositoryImpl) CompleteJob(ctx context.Context, jobID string) (*queries.Job, error) {
	q := queries.New(r.db.pool)

	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	job, err := q.CompleteJob(ctx, pgtype.UUID{Bytes: jobUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to complete job: %w", err)
	}

	return job, nil
}

// FailJob marks a job as failed with an error message and sets the finished timestamp.
func (r *JobRepositoryImpl) FailJob(ctx context.Context, jobID string, errorMsg string) (*queries.Job, error) {
	q := queries.New(r.db.pool)

	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	job, err := q.FailJob(ctx, queries.FailJobParams{
		ID:    pgtype.UUID{Bytes: jobUUID, Valid: true},
		Error: pgtype.Text{String: errorMsg, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fail job: %w", err)
	}

	return job, nil
}

// GetPendingJobs retrieves a limited number of pending jobs.
func (r *JobRepositoryImpl) GetPendingJobs(ctx context.Context, limit int) ([]*queries.Job, error) {
	q := queries.New(r.db.pool)

	jobs, err := q.GetPendingJobs(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}

	return jobs, nil
}

// DeleteJob deletes a job from the database.
func (r *JobRepositoryImpl) DeleteJob(ctx context.Context, jobID string) error {
	q := queries.New(r.db.pool)

	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	err = q.DeleteJob(ctx, pgtype.UUID{Bytes: jobUUID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

// DeleteJobsByImageID deletes all jobs for a specific image.
func (r *JobRepositoryImpl) DeleteJobsByImageID(ctx context.Context, imageID string) error {
	q := queries.New(r.db.pool)

	imageUUID, err := uuid.Parse(imageID)
	if err != nil {
		return fmt.Errorf("invalid image ID: %w", err)
	}

	err = q.DeleteJobsByImageID(ctx, pgtype.UUID{Bytes: imageUUID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to delete jobs: %w", err)
	}

	return nil
}
