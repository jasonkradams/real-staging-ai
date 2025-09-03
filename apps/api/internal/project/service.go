package project

import "context"

//go:generate go run github.com/matryer/moq@v0.5.3 -out service_mock.go . Service

type Service interface {
	CreateProject(ctx context.Context, req *CreateRequest) (*Project, error)
	CreateProjectWithUpload(ctx context.Context, req *CreateRequest, filename, contentType string, fileSize int64) (*WithUploadURL, error)
	GetProjectsByUser(ctx context.Context, userID string) ([]Project, error)
	GetProjectByID(ctx context.Context, projectID, userID string) (*Project, error)
	UpdateProject(ctx context.Context, projectID, userID, newName string) (*Project, error)
	DeleteProject(ctx context.Context, projectID, userID string) error
	GetProjectStats(ctx context.Context, userID string) (*ProjectStats, error)
}
