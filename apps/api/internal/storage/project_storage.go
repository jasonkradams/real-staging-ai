package storage

import (
	"context"
	"fmt"

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
		RETURNING id, created_at
	`
	// For now, we'll use a hardcoded user_id.
	// We'll get this from the context in the future.
	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" // from seed data

	err := s.db.pool.QueryRow(ctx, query, p.Name, userID).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to create project: %w", err)
	}

	return p, nil
}

// GetProjects retrieves all projects from the database.
func (s *ProjectStorage) GetProjects(ctx context.Context) ([]project.Project, error) {
	query := `
		SELECT id, name, created_at
		FROM projects
	`
	rows, err := s.db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unable to get projects: %w", err)
	}
	defer rows.Close()

	var projects []project.Project
	for rows.Next() {
		var p project.Project
		err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}
