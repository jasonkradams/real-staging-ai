package project

import (
	"context"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out repository_mock.go . Repository

// Repository defines the interface for project data access operations.
type Repository interface {
	// CreateProject creates a new project in the database.
	CreateProject(ctx context.Context, p *Project, userID string) (*Project, error)

	// GetProjects retrieves all projects from the database.
	GetProjects(ctx context.Context) ([]Project, error)

	// GetProjectsByUserID retrieves all projects for a specific user.
	GetProjectsByUserID(ctx context.Context, userID string) ([]Project, error)

	// GetProjectByID retrieves a specific project by its ID.
	GetProjectByID(ctx context.Context, projectID string) (*Project, error)

	// GetProjectByIDAndUserID retrieves a specific project by its ID and user ID.
	GetProjectByIDAndUserID(ctx context.Context, projectID, userID string) (*Project, error)

	// UpdateProject updates an existing project's name.
	UpdateProject(ctx context.Context, projectID, name string) (*Project, error)

	// UpdateProjectByUserID updates an existing project's name with user ownership verification.
	UpdateProjectByUserID(ctx context.Context, projectID, userID, name string) (*Project, error)

	// DeleteProject deletes a project from the database.
	DeleteProject(ctx context.Context, projectID string) error

	// DeleteProjectByUserID deletes a project from the database with user ownership verification.
	DeleteProjectByUserID(ctx context.Context, projectID, userID string) error

	// CountProjectsByUserID returns the number of projects for a specific user.
	CountProjectsByUserID(ctx context.Context, userID string) (int64, error)
}
