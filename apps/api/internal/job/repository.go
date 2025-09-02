package job

import (
	"context"

	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out repository_mock.go . Repository

// Repository defines the interface for job data access operations.
type Repository interface {
	// CreateJob creates a new job in the database.
	CreateJob(ctx context.Context, imageID string, jobType string, payloadJSON []byte) (*queries.Job, error)

	// GetJobByID retrieves a specific job by its ID.
	GetJobByID(ctx context.Context, jobID string) (*queries.Job, error)

	// GetJobsByImageID retrieves all jobs for a specific image.
	GetJobsByImageID(ctx context.Context, imageID string) ([]*queries.Job, error)

	// UpdateJobStatus updates a job's status.
	UpdateJobStatus(ctx context.Context, jobID string, status string) (*queries.Job, error)

	// StartJob marks a job as processing and sets the started timestamp.
	StartJob(ctx context.Context, jobID string) (*queries.Job, error)

	// CompleteJob marks a job as completed and sets the finished timestamp.
	CompleteJob(ctx context.Context, jobID string) (*queries.Job, error)

	// FailJob marks a job as failed with an error message and sets the finished timestamp.
	FailJob(ctx context.Context, jobID string, errorMsg string) (*queries.Job, error)

	// GetPendingJobs retrieves a limited number of pending jobs.
	GetPendingJobs(ctx context.Context, limit int) ([]*queries.Job, error)

	// DeleteJob deletes a job from the database.
	DeleteJob(ctx context.Context, jobID string) error

	// DeleteJobsByImageID deletes all jobs for a specific image.
	DeleteJobsByImageID(ctx context.Context, imageID string) error
}
