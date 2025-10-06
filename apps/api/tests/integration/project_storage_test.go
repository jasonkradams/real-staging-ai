//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/virtual-staging-ai/api/internal/project"
)

func TestProjectStorage_CreateProject(t *testing.T) {
	// Setup
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	TruncateAllTables(ctx, db.Pool())
	SeedDatabase(ctx, db.Pool())

	projectStorage := project.NewDefaultRepository(db)

	// Create a new project
	newProject := &project.Project{
		Name: "Test Project",
	}
	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" // Use seeded user ID
	createdProject, err := projectStorage.CreateProject(ctx, newProject, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, createdProject)
	assert.NotEmpty(t, createdProject.ID)
	assert.Equal(t, "Test Project", createdProject.Name)
}

func TestProjectStorage_GetProjects(t *testing.T) {
	// Setup
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	TruncateAllTables(ctx, db.Pool())
	SeedDatabase(ctx, db.Pool())

	projectStorage := project.NewDefaultRepository(db)

	// We have seeded the database with one project, so we expect to get one project back.
	projects, err := projectStorage.GetProjects(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, projects)
	assert.Len(t, projects, 1)
}
