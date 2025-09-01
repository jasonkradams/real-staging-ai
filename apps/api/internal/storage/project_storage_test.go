//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/virtual-staging-ai/api/internal/project"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/testutil"
)

func TestProjectStorage_CreateProject(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	assert.NoError(t, err)
	defer db.Close()

	testutil.TruncateTables(t, db.GetPool())
	testutil.SeedTables(t, db.GetPool())

	projectStorage := storage.NewProjectStorage(db)

	// Create a new project
	newProject := &project.Project{
		Name: "Test Project",
	}
	createdProject, err := projectStorage.CreateProject(ctx, newProject)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, createdProject)
	assert.NotEmpty(t, createdProject.ID)
	assert.Equal(t, "Test Project", createdProject.Name)
}

func TestProjectStorage_GetProjects(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	assert.NoError(t, err)
	defer db.Close()

	testutil.TruncateTables(t, db.GetPool())
	testutil.SeedTables(t, db.GetPool())

	projectStorage := storage.NewProjectStorage(db)

	// We have seeded the database with one project, so we expect to get one project back.
	projects, err := projectStorage.GetProjects(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, projects)
	assert.Len(t, projects, 1)
}
