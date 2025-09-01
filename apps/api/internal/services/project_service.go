package services

import (
	"context"
	"fmt"

	"github.com/virtual-staging-ai/api/internal/project"
	"github.com/virtual-staging-ai/api/internal/storage"
)

// ProjectService provides business logic for project operations.
// It demonstrates proper dependency injection for testability.
type ProjectService struct {
	projectRepo storage.ProjectRepository
	s3Service   storage.S3Service
}

// NewProjectService creates a new ProjectService with the provided dependencies.
func NewProjectService(projectRepo storage.ProjectRepository, s3Service storage.S3Service) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		s3Service:   s3Service,
	}
}

// CreateProjectRequest represents the input for creating a project.
type CreateProjectRequest struct {
	Name   string `json:"name" validate:"required,min=1,max=100"`
	UserID string `json:"user_id" validate:"required"`
}

// ProjectWithUploadURL represents a project with an associated upload URL.
type ProjectWithUploadURL struct {
	Project   *project.Project `json:"project"`
	UploadURL string           `json:"upload_url,omitempty"`
	FileKey   string           `json:"file_key,omitempty"`
}

// CreateProject creates a new project and optionally generates an upload URL.
func (s *ProjectService) CreateProject(ctx context.Context, req *CreateProjectRequest) (*project.Project, error) {
	// Validate input
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}
	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Create project entity
	proj := &project.Project{
		Name: req.Name,
	}

	// Create project in repository
	createdProject, err := s.projectRepo.CreateProject(ctx, proj, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return createdProject, nil
}

// CreateProjectWithUpload creates a project and generates a presigned upload URL.
func (s *ProjectService) CreateProjectWithUpload(ctx context.Context, req *CreateProjectRequest, filename, contentType string, fileSize int64) (*ProjectWithUploadURL, error) {
	// Create the project first
	createdProject, err := s.CreateProject(ctx, req)
	if err != nil {
		return nil, err
	}

	// Generate upload URL
	uploadResult, err := s.s3Service.GeneratePresignedUploadURL(ctx, req.UserID, filename, contentType, fileSize)
	if err != nil {
		// Project was created but upload URL failed - in a real system you might want to handle this differently
		return &ProjectWithUploadURL{
			Project: createdProject,
		}, fmt.Errorf("project created but failed to generate upload URL: %w", err)
	}

	return &ProjectWithUploadURL{
		Project:   createdProject,
		UploadURL: uploadResult.UploadURL,
		FileKey:   uploadResult.FileKey,
	}, nil
}

// GetProjectsByUser retrieves all projects for a specific user.
func (s *ProjectService) GetProjectsByUser(ctx context.Context, userID string) ([]project.Project, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	projects, err := s.projectRepo.GetProjectsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects for user: %w", err)
	}

	return projects, nil
}

// GetProjectByID retrieves a project by its ID, ensuring the user owns it.
func (s *ProjectService) GetProjectByID(ctx context.Context, projectID, userID string) (*project.Project, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	proj, err := s.projectRepo.GetProjectByIDAndUserID(ctx, projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return proj, nil
}

// UpdateProject updates a project's name, ensuring the user owns it.
func (s *ProjectService) UpdateProject(ctx context.Context, projectID, userID, newName string) (*project.Project, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if newName == "" {
		return nil, fmt.Errorf("project name is required")
	}

	updatedProject, err := s.projectRepo.UpdateProjectByUserID(ctx, projectID, userID, newName)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return updatedProject, nil
}

// DeleteProject deletes a project, ensuring the user owns it.
func (s *ProjectService) DeleteProject(ctx context.Context, projectID, userID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	err := s.projectRepo.DeleteProjectByUserID(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// GetProjectStats returns statistics about a user's projects.
func (s *ProjectService) GetProjectStats(ctx context.Context, userID string) (*ProjectStats, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	count, err := s.projectRepo.CountProjectsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	return &ProjectStats{
		TotalProjects: count,
	}, nil
}

// ProjectStats represents statistics about a user's projects.
type ProjectStats struct {
	TotalProjects int64 `json:"total_projects"`
}
