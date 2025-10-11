package project

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/real-staging-ai/api/internal/storage"
)

// DefaultRepository handles the database operations for projects.
type DefaultRepository struct {
	db storage.Database
}

// Ensure DefaultRepository implements Repository interface.
var _ Repository = (*DefaultRepository)(nil)

// NewDefaultRepository creates a new DefaultRepository instance.
func NewDefaultRepository(db storage.Database) *DefaultRepository {
	return &DefaultRepository{db: db}
}

// CreateProject creates a new project in the database.
func (s *DefaultRepository) CreateProject(ctx context.Context, p *Project, userID string) (*Project, error) {
	query := `
		INSERT INTO projects (name, user_id)
		VALUES ($1, $2)
		RETURNING id, user_id, created_at
	`
	// Use the provided userID parameter
	if userID == "" {
		// Fallback to hardcoded user_id for backward compatibility
		userID = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" // from seed data
	}

	err := s.db.QueryRow(ctx, query, p.Name, userID).Scan(&p.ID, &p.UserID, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to create project: %w", err)
	}

	// Removed self-assignment of p.Name

	return p, nil
}

// GetProjects retrieves all projects from the database.
// TODO: Filter by user_id when auth middleware is implemented.
func (s *DefaultRepository) GetProjects(ctx context.Context) ([]Project, error) {
	query := `
		SELECT id, name, user_id, created_at
		FROM projects
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unable to get projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.Name, &p.UserID, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over project rows: %w", err)
	}

	return projects, nil
}

// GetProjectsByUserID retrieves all projects for a specific user.
func (s *DefaultRepository) GetProjectsByUserID(ctx context.Context, userID string) ([]Project, error) {
	query := `
		SELECT id, name, user_id, created_at
		FROM projects
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to get projects for user: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.Name, &p.UserID, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over project rows: %w", err)
	}

	return projects, nil
}

// GetProjectByIDAndUserID retrieves a specific project by its ID and user ID.
func (s *DefaultRepository) GetProjectByIDAndUserID(ctx context.Context, projectID, userID string) (*Project, error) {
	query := `
		SELECT id, name, user_id, created_at
		FROM projects
		WHERE id = $1 AND user_id = $2
	`

	var p Project
	err := s.db.QueryRow(ctx, query, projectID, userID).Scan(&p.ID, &p.Name, &p.UserID, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get project by ID and user ID: %w", err)
	}

	return &p, nil
}

// DeleteProjectByUserID deletes a project from the database with user ownership verification.
func (s *DefaultRepository) DeleteProjectByUserID(ctx context.Context, projectID, userID string) error {
	query := `
		DELETE FROM projects
		WHERE id = $1 AND user_id = $2
	`

	result, err := s.db.Exec(ctx, query, projectID, userID)
	if err != nil {
		return fmt.Errorf("unable to delete project: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// CountProjectsByUserID returns the number of projects for a specific user.
func (s *DefaultRepository) CountProjectsByUserID(ctx context.Context, userID string) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM projects
		WHERE user_id = $1
	`

	var count int64
	err := s.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to count projects for user: %w", err)
	}

	return count, nil
}

// GetProjectByID retrieves a specific project by its ID.
// TODO: Add user_id filtering when auth middleware is implemented.
func (s *DefaultRepository) GetProjectByID(ctx context.Context, projectID string) (*Project, error) {
	query := `
		SELECT id, name, user_id, created_at
		FROM projects
		WHERE id = $1
	`

	var p Project
	err := s.db.QueryRow(ctx, query, projectID).Scan(&p.ID, &p.Name, &p.UserID, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get project by ID: %w", err)
	}

	return &p, nil
}

// UpdateProjectByUserID updates an existing project's name with user ownership verification.
func (s *DefaultRepository) UpdateProjectByUserID(ctx context.Context, projectID, userID, name string) (*Project, error) {
	query := `
		UPDATE projects
		SET name = $3
		WHERE id = $1 AND user_id = $2
		RETURNING id, name, user_id, created_at
	`

	var p Project
	err := s.db.QueryRow(ctx, query, projectID, userID, name).Scan(&p.ID, &p.Name, &p.UserID, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to update project: %w", err)
	}

	return &p, nil
}

// UpdateProject updates an existing project's name.
// TODO: Add user_id filtering when auth middleware is implemented.
func (s *DefaultRepository) UpdateProject(ctx context.Context, projectID, name string) (*Project, error) {
	query := `
		UPDATE projects
		SET name = $2
		WHERE id = $1
		RETURNING id, name, user_id, created_at
	`

	var p Project
	err := s.db.QueryRow(ctx, query, projectID, name).Scan(&p.ID, &p.Name, &p.UserID, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to update project: %w", err)
	}

	return &p, nil
}

// DeleteProject deletes a project from the database.
// TODO: Add user_id filtering when auth middleware is implemented.
func (s *DefaultRepository) DeleteProject(ctx context.Context, projectID string) error {
	query := `
		DELETE FROM projects
		WHERE id = $1
	`

	result, err := s.db.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("unable to delete project: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
