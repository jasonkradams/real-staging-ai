package project

import "context"

//go:generate go run github.com/matryer/moq@v0.5.3 -out storage_mock.go . Storage

type Storage interface {
	CreateProject(ctx context.Context, p *Project, userID string) (*Project, error)
	GetProjects(ctx context.Context) ([]Project, error)
	GetProjectsByUserID(ctx context.Context, userID string) ([]Project, error)
	GetProjectByIDAndUserID(ctx context.Context, projectID, userID string) (*Project, error)
	DeleteProjectByUserID(ctx context.Context, projectID, userID string) error
	CountProjectsByUserID(ctx context.Context, userID string) (int64, error)
	GetProjectByID(ctx context.Context, projectID string) (*Project, error)
	UpdateProjectByUserID(ctx context.Context, projectID, userID, name string) (*Project, error)
	UpdateProject(ctx context.Context, projectID, name string) (*Project, error)
	DeleteProject(ctx context.Context, projectID string) error
}
