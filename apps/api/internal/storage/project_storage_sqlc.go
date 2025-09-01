package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/virtual-staging-ai/api/internal/project"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// ProjectStorageSQLc handles the database operations for projects using sqlc-generated queries.
type ProjectStorageSQLc struct {
	queries *queries.Queries
}

// Ensure ProjectStorageSQLc implements ProjectRepository interface.
var _ ProjectRepository = (*ProjectStorageSQLc)(nil)

// NewProjectStorageSQLc creates a new ProjectStorageSQLc instance.
func NewProjectStorageSQLc(db *DB) *ProjectStorageSQLc {
	return &ProjectStorageSQLc{
		queries: queries.New(db.pool),
	}
}

// CreateProject creates a new project in the database.
func (s *ProjectStorageSQLc) CreateProject(ctx context.Context, p *project.Project, userID string) (*project.Project, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{}
	err = userUUIDType.Scan(userUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID to pgtype.UUID: %w", err)
	}

	params := queries.CreateProjectParams{
		Name:   p.Name,
		UserID: userUUIDType,
	}

	result, err := s.queries.CreateProject(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to create project: %w", err)
	}

	// Convert back to project.Project
	createdProject := &project.Project{
		ID:        uuid.UUID(result.ID.Bytes).String(),
		Name:      result.Name,
		UserID:    uuid.UUID(result.UserID.Bytes).String(),
		CreatedAt: result.CreatedAt.Time,
	}

	return createdProject, nil
}

// GetProjects retrieves all projects from the database.
// TODO: Filter by user_id when auth middleware is implemented.
func (s *ProjectStorageSQLc) GetProjects(ctx context.Context) ([]project.Project, error) {
	results, err := s.queries.GetAllProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get projects: %w", err)
	}

	projects := make([]project.Project, 0, len(results))
	for _, result := range results {
		p := project.Project{
			ID:        uuid.UUID(result.ID.Bytes).String(),
			Name:      result.Name,
			UserID:    uuid.UUID(result.UserID.Bytes).String(),
			CreatedAt: result.CreatedAt.Time,
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// GetProjectsByUserID retrieves all projects for a specific user.
func (s *ProjectStorageSQLc) GetProjectsByUserID(ctx context.Context, userID string) ([]project.Project, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{}
	err = userUUIDType.Scan(userUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID to pgtype.UUID: %w", err)
	}

	results, err := s.queries.GetProjectsByUserID(ctx, userUUIDType)
	if err != nil {
		return nil, fmt.Errorf("unable to get projects for user: %w", err)
	}

	projects := make([]project.Project, 0, len(results))
	for _, result := range results {
		p := project.Project{
			ID:        uuid.UUID(result.ID.Bytes).String(),
			Name:      result.Name,
			UserID:    uuid.UUID(result.UserID.Bytes).String(),
			CreatedAt: result.CreatedAt.Time,
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// GetProjectByID retrieves a specific project by its ID.
// TODO: Add user_id filtering when auth middleware is implemented.
func (s *ProjectStorageSQLc) GetProjectByID(ctx context.Context, projectID string) (*project.Project, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID format: %w", err)
	}

	projectUUIDType := pgtype.UUID{}
	err = projectUUIDType.Scan(projectUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert project ID to pgtype.UUID: %w", err)
	}

	result, err := s.queries.GetProjectByID(ctx, projectUUIDType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get project by ID: %w", err)
	}

	p := &project.Project{
		ID:        uuid.UUID(result.ID.Bytes).String(),
		Name:      result.Name,
		UserID:    uuid.UUID(result.UserID.Bytes).String(),
		CreatedAt: result.CreatedAt.Time,
	}

	return p, nil
}

// GetProjectByIDAndUserID retrieves a specific project by its ID and user ID.
func (s *ProjectStorageSQLc) GetProjectByIDAndUserID(ctx context.Context, projectID, userID string) (*project.Project, error) {
	// For now, we'll get the project and check the user ID manually
	// This could be optimized with a dedicated query if needed
	project, err := s.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project.UserID != userID {
		return nil, pgx.ErrNoRows // Project not found for this user
	}

	return project, nil
}

// UpdateProject updates an existing project's name.
// TODO: Add user_id filtering when auth middleware is implemented.
func (s *ProjectStorageSQLc) UpdateProject(ctx context.Context, projectID, name string) (*project.Project, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID format: %w", err)
	}

	projectUUIDType := pgtype.UUID{}
	err = projectUUIDType.Scan(projectUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert project ID to pgtype.UUID: %w", err)
	}

	params := queries.UpdateProjectParams{
		ID:   projectUUIDType,
		Name: name,
	}

	result, err := s.queries.UpdateProject(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to update project: %w", err)
	}

	p := &project.Project{
		ID:        uuid.UUID(result.ID.Bytes).String(),
		Name:      result.Name,
		UserID:    uuid.UUID(result.UserID.Bytes).String(),
		CreatedAt: result.CreatedAt.Time,
	}

	return p, nil
}

// UpdateProjectByUserID updates an existing project's name with user ownership verification.
func (s *ProjectStorageSQLc) UpdateProjectByUserID(ctx context.Context, projectID, userID, name string) (*project.Project, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID format: %w", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	projectUUIDType := pgtype.UUID{}
	err = projectUUIDType.Scan(projectUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert project ID to pgtype.UUID: %w", err)
	}

	userUUIDType := pgtype.UUID{}
	err = userUUIDType.Scan(userUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID to pgtype.UUID: %w", err)
	}

	params := queries.UpdateProjectByUserIDParams{
		ID:     projectUUIDType,
		UserID: userUUIDType,
		Name:   name,
	}

	result, err := s.queries.UpdateProjectByUserID(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to update project: %w", err)
	}

	p := &project.Project{
		ID:        uuid.UUID(result.ID.Bytes).String(),
		Name:      result.Name,
		UserID:    uuid.UUID(result.UserID.Bytes).String(),
		CreatedAt: result.CreatedAt.Time,
	}

	return p, nil
}

// DeleteProject deletes a project from the database.
// TODO: Add user_id filtering when auth middleware is implemented.
func (s *ProjectStorageSQLc) DeleteProject(ctx context.Context, projectID string) error {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID format: %w", err)
	}

	projectUUIDType := pgtype.UUID{}
	err = projectUUIDType.Scan(projectUUID.String())
	if err != nil {
		return fmt.Errorf("failed to convert project ID to pgtype.UUID: %w", err)
	}

	err = s.queries.DeleteProject(ctx, projectUUIDType)
	if err != nil {
		return fmt.Errorf("unable to delete project: %w", err)
	}

	// Note: sqlc's :exec queries don't return the number of affected rows
	// We would need to check if the project exists first if we want to return pgx.ErrNoRows
	// For now, we'll assume the delete was successful

	return nil
}

// DeleteProjectByUserID deletes a project from the database with user ownership verification.
func (s *ProjectStorageSQLc) DeleteProjectByUserID(ctx context.Context, projectID, userID string) error {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID format: %w", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	projectUUIDType := pgtype.UUID{}
	err = projectUUIDType.Scan(projectUUID.String())
	if err != nil {
		return fmt.Errorf("failed to convert project ID to pgtype.UUID: %w", err)
	}

	userUUIDType := pgtype.UUID{}
	err = userUUIDType.Scan(userUUID.String())
	if err != nil {
		return fmt.Errorf("failed to convert user ID to pgtype.UUID: %w", err)
	}

	params := queries.DeleteProjectByUserIDParams{
		ID:     projectUUIDType,
		UserID: userUUIDType,
	}

	err = s.queries.DeleteProjectByUserID(ctx, params)
	if err != nil {
		return fmt.Errorf("unable to delete project: %w", err)
	}

	return nil
}

// CountProjectsByUserID returns the number of projects for a specific user.
func (s *ProjectStorageSQLc) CountProjectsByUserID(ctx context.Context, userID string) (int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{}
	err = userUUIDType.Scan(userUUID.String())
	if err != nil {
		return 0, fmt.Errorf("failed to convert user ID to pgtype.UUID: %w", err)
	}

	count, err := s.queries.CountProjectsByUserID(ctx, userUUIDType)
	if err != nil {
		return 0, fmt.Errorf("unable to count projects for user: %w", err)
	}

	return count, nil
}
