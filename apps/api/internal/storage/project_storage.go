package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/virtual-staging-ai/api/internal/project"
)

// ProjectStorage handles the database operations for projects.
type ProjectStorage struct {
	db *DB
}

// NewProjectStorage creates a new ProjectStorage instance.
func NewProjectStorage(db *DB) *ProjectStorage {
	return &ProjectStorage{db: db}
}

// CreateProject creates a new project in the database.
func (s *ProjectStorage) CreateProject(ctx context.Context, p *project.Project) (*project.Project, error) {
	query := `
		INSERT INTO projects (name, user_id)
		VALUES ($1, $2)
		RETURNING id, user_id, created_at
	`
	// For now, we'll use a hardcoded user_id.
	// We'll get this from the context in the future when auth middleware is implemented.
	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" // from seed data

	err := s.db.pool.QueryRow(ctx, query, p.Name, userID).Scan(&p.ID, &p.UserID, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to create project: %w", err)
	}

	return p, nil
}

// GetProjects retrieves all projects from the database.
// TODO: Filter by user_id when auth middleware is implemented.
func (s *ProjectStorage) GetProjects(ctx context.Context) ([]project.Project, error) {
	query := `
		SELECT id, name, user_id, created_at
		FROM projects
		ORDER BY created_at DESC
	`
	rows, err := s.db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unable to get projects: %w", err)
	}
	defer rows.Close()

	var projects []project.Project
	for rows.Next() {
		var p project.Project
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

// GetProjectByID retrieves a specific project by its ID.
// TODO: Add user_id filtering when auth middleware is implemented.
func (s *ProjectStorage) GetProjectByID(ctx context.Context, projectID string) (*project.Project, error) {
	query := `
		SELECT id, name, user_id, created_at
		FROM projects
		WHERE id = $1
	`

	var p project.Project
	err := s.db.pool.QueryRow(ctx, query, projectID).Scan(&p.ID, &p.Name, &p.UserID, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get project by ID: %w", err)
	}

	return &p, nil
}

// UpdateProject updates an existing project's name.
// TODO: Add user_id filtering when auth middleware is implemented.
func (s *ProjectStorage) UpdateProject(ctx context.Context, projectID, name string) (*project.Project, error) {
	query := `
		UPDATE projects
		SET name = $1
		WHERE id = $2
		RETURNING id, name, user_id, created_at
	`

	var p project.Project
	err := s.db.pool.QueryRow(ctx, query, name, projectID).Scan(&p.ID, &p.Name, &p.UserID, &p.CreatedAt)
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
func (s *ProjectStorage) DeleteProject(ctx context.Context, projectID string) error {
	query := `
		DELETE FROM projects
		WHERE id = $1
	`

	result, err := s.db.pool.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("unable to delete project: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
