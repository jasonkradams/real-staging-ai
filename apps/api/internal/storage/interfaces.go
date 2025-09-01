package storage

//go:generate go run go.uber.org/mock/mockgen@latest -package=mocks -destination=./mocks/project_repository_mock.go github.com/virtual-staging-ai/api/internal/storage ProjectRepository
//go:generate go run go.uber.org/mock/mockgen@latest -package=mocks -destination=./mocks/s3_service_mock.go github.com/virtual-staging-ai/api/internal/storage S3Service
//go:generate go run go.uber.org/mock/mockgen@latest -package=mocks -destination=./mocks/image_repository_mock.go github.com/virtual-staging-ai/api/internal/storage ImageRepository
//go:generate go run go.uber.org/mock/mockgen@latest -package=mocks -destination=./mocks/job_repository_mock.go github.com/virtual-staging-ai/api/internal/storage JobRepository
//go:generate go run go.uber.org/mock/mockgen@latest -package=mocks -destination=./mocks/user_repository_mock.go github.com/virtual-staging-ai/api/internal/storage UserRepository

import (
	"context"

	"github.com/virtual-staging-ai/api/internal/project"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// ProjectRepository defines the interface for project data access operations.
type ProjectRepository interface {
	// CreateProject creates a new project in the database.
	CreateProject(ctx context.Context, p *project.Project, userID string) (*project.Project, error)

	// GetProjects retrieves all projects from the database.
	GetProjects(ctx context.Context) ([]project.Project, error)

	// GetProjectsByUserID retrieves all projects for a specific user.
	GetProjectsByUserID(ctx context.Context, userID string) ([]project.Project, error)

	// GetProjectByID retrieves a specific project by its ID.
	GetProjectByID(ctx context.Context, projectID string) (*project.Project, error)

	// GetProjectByIDAndUserID retrieves a specific project by its ID and user ID.
	GetProjectByIDAndUserID(ctx context.Context, projectID, userID string) (*project.Project, error)

	// UpdateProject updates an existing project's name.
	UpdateProject(ctx context.Context, projectID, name string) (*project.Project, error)

	// UpdateProjectByUserID updates an existing project's name with user ownership verification.
	UpdateProjectByUserID(ctx context.Context, projectID, userID, name string) (*project.Project, error)

	// DeleteProject deletes a project from the database.
	DeleteProject(ctx context.Context, projectID string) error

	// DeleteProjectByUserID deletes a project from the database with user ownership verification.
	DeleteProjectByUserID(ctx context.Context, projectID, userID string) error

	// CountProjectsByUserID returns the number of projects for a specific user.
	CountProjectsByUserID(ctx context.Context, userID string) (int64, error)
}

// S3Service defines the interface for S3 storage operations.
type S3Service interface {
	// GeneratePresignedUploadURL generates a presigned URL for uploading a file to S3.
	GeneratePresignedUploadURL(ctx context.Context, userID, filename, contentType string, fileSize int64) (*PresignedUploadResult, error)

	// GetFileURL returns the public URL for a file in S3.
	GetFileURL(fileKey string) string

	// DeleteFile deletes a file from S3.
	DeleteFile(ctx context.Context, fileKey string) error

	// HeadFile checks if a file exists in S3 and returns its metadata.
	HeadFile(ctx context.Context, fileKey string) (interface{}, error)
}

// ImageRepository defines the interface for image data access operations.
type ImageRepository interface {
	// CreateImage creates a new image in the database.
	CreateImage(ctx context.Context, projectID string, originalURL string, roomType, style *string, seed *int64) (*queries.Image, error)

	// GetImageByID retrieves a specific image by its ID.
	GetImageByID(ctx context.Context, imageID string) (*queries.Image, error)

	// GetImagesByProjectID retrieves all images for a specific project.
	GetImagesByProjectID(ctx context.Context, projectID string) ([]*queries.Image, error)

	// UpdateImageStatus updates an image's processing status.
	UpdateImageStatus(ctx context.Context, imageID string, status string) (*queries.Image, error)

	// UpdateImageWithStagedURL updates an image with the staged URL and status.
	UpdateImageWithStagedURL(ctx context.Context, imageID string, stagedURL string, status string) (*queries.Image, error)

	// UpdateImageWithError updates an image with an error status and message.
	UpdateImageWithError(ctx context.Context, imageID string, errorMsg string) (*queries.Image, error)

	// DeleteImage deletes an image from the database.
	DeleteImage(ctx context.Context, imageID string) error

	// DeleteImagesByProjectID deletes all images for a specific project.
	DeleteImagesByProjectID(ctx context.Context, projectID string) error
}

// JobRepository defines the interface for job data access operations.
type JobRepository interface {
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

// UserRepository defines the interface for user data access operations.
type UserRepository interface {
	// CreateUser creates a new user in the database.
	CreateUser(ctx context.Context, auth0Sub, stripeCustomerID, role string) (*queries.User, error)

	// GetUserByID retrieves a user by their ID.
	GetUserByID(ctx context.Context, userID string) (*queries.User, error)

	// GetUserByAuth0Sub retrieves a user by their Auth0 subject ID.
	GetUserByAuth0Sub(ctx context.Context, auth0Sub string) (*queries.User, error)

	// GetUserByStripeCustomerID retrieves a user by their Stripe customer ID.
	GetUserByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*queries.User, error)

	// UpdateUserStripeCustomerID updates a user's Stripe customer ID.
	UpdateUserStripeCustomerID(ctx context.Context, userID, stripeCustomerID string) (*queries.User, error)

	// UpdateUserRole updates a user's role.
	UpdateUserRole(ctx context.Context, userID, role string) (*queries.User, error)

	// DeleteUser deletes a user from the database.
	DeleteUser(ctx context.Context, userID string) error

	// ListUsers retrieves a paginated list of users.
	ListUsers(ctx context.Context, limit, offset int) ([]*queries.User, error)

	// CountUsers returns the total number of users.
	CountUsers(ctx context.Context) (int64, error)
}

// Querier represents the interface for database queries (from sqlc).
type Querier interface {
	queries.Querier
}
