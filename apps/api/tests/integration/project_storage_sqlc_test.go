//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/real-staging-ai/api/internal/project"
	"github.com/real-staging-ai/api/internal/storage"
)

func TestProjectStorageSQLc_CreateProject(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	truncateTables(t, db.Pool())
	seedTables(t, db.Pool())

	storageInstance := project.NewDefaultStorageSQLc(db)
	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" // from seed data

	testCases := []struct {
		name         string
		project      *project.Project
		userID       string
		expectError  bool
		expectedName string
	}{
		{
			name: "success: valid project creation",
			project: &project.Project{
				Name: "Test Project",
			},
			userID:       userID,
			expectError:  false,
			expectedName: "Test Project",
		},
		{
			name: "fail: invalid user ID format",
			project: &project.Project{
				Name: "Test Project",
			},
			userID:      "invalid-uuid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdProject, err := storageInstance.CreateProject(ctx, tc.project, tc.userID)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, createdProject.ID)
			assert.Equal(t, tc.expectedName, createdProject.Name)
			assert.Equal(t, tc.userID, createdProject.UserID)
			assert.NotZero(t, createdProject.CreatedAt)
		})
	}
}

func TestProjectStorageSQLc_GetProjects(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	truncateTables(t, db.Pool())
	seedTables(t, db.Pool())

	storageInstance := project.NewDefaultStorageSQLc(db)

	projects, err := storageInstance.GetProjects(ctx)
	require.NoError(t, err)

	assert.Len(t, projects, 1)
	assert.Equal(t, "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", projects[0].ID)
	assert.Equal(t, "Test Project 1", projects[0].Name)
	assert.Equal(t, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", projects[0].UserID)
}

func TestProjectStorageSQLc_GetProjectsByUserID(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	truncateTables(t, db.Pool())
	seedTables(t, db.Pool())

	storageInstance := project.NewDefaultStorageSQLc(db)
	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

	testCases := []struct {
		name            string
		userID          string
		expectError     bool
		expectedCount   int
		expectedProject string
	}{
		{
			name:            "success: get projects for existing user",
			userID:          userID,
			expectError:     false,
			expectedCount:   1,
			expectedProject: "Test Project 1",
		},
		{
			name:          "success: get projects for non-existent user",
			userID:        "550e8400-e29b-41d4-a716-446655440000",
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:        "fail: invalid user ID format",
			userID:      "invalid-uuid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projects, err := storageInstance.GetProjectsByUserID(ctx, tc.userID)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, projects, tc.expectedCount)

			if tc.expectedCount > 0 {
				assert.Equal(t, tc.expectedProject, projects[0].Name)
				assert.Equal(t, tc.userID, projects[0].UserID)
			}
		})
	}
}

func TestProjectStorageSQLc_GetProjectByID(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	truncateTables(t, db.Pool())
	seedTables(t, db.Pool())

	storageInstance := project.NewDefaultStorageSQLc(db)

	testCases := []struct {
		name         string
		projectID    string
		expectError  bool
		errorType    error
		expectedName string
	}{
		{
			name:         "success: get existing project",
			projectID:    "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			expectError:  false,
			expectedName: "Test Project 1",
		},
		{
			name:        "fail: project not found",
			projectID:   "550e8400-e29b-41d4-a716-446655440000",
			expectError: true,
		},
		{
			name:        "fail: invalid project ID format",
			projectID:   "invalid-uuid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			project, err := storageInstance.GetProjectByID(ctx, tc.projectID)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.projectID, project.ID)
			assert.Equal(t, tc.expectedName, project.Name)
			assert.NotEmpty(t, project.UserID)
		})
	}
}

func TestProjectStorageSQLc_UpdateProject(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	truncateTables(t, db.Pool())
	seedTables(t, db.Pool())

	storageInstance := project.NewDefaultStorageSQLc(db)

	testCases := []struct {
		name         string
		projectID    string
		newName      string
		expectError  bool
		expectedName string
	}{
		{
			name:         "success: update existing project",
			projectID:    "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			newName:      "Updated Project Name",
			expectError:  false,
			expectedName: "Updated Project Name",
		},
		{
			name:        "fail: project not found",
			projectID:   "550e8400-e29b-41d4-a716-446655440000",
			newName:     "Updated Name",
			expectError: true,
		},
		{
			name:        "fail: invalid project ID format",
			projectID:   "invalid-uuid",
			newName:     "Updated Name",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updatedProject, err := storageInstance.UpdateProject(ctx, tc.projectID, tc.newName)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.projectID, updatedProject.ID)
			assert.Equal(t, tc.expectedName, updatedProject.Name)
		})
	}
}

func TestProjectStorageSQLc_UpdateProjectByUserID(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	truncateTables(t, db.Pool())
	seedTables(t, db.Pool())

	storageInstance := project.NewDefaultStorageSQLc(db)
	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

	testCases := []struct {
		name         string
		projectID    string
		userID       string
		newName      string
		expectError  bool
		expectedName string
	}{
		{
			name:         "success: update project by correct user",
			projectID:    "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			userID:       userID,
			newName:      "Updated by User",
			expectError:  false,
			expectedName: "Updated by User",
		},
		{
			name:        "fail: update project by wrong user",
			projectID:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			userID:      "550e8400-e29b-41d4-a716-446655440000",
			newName:     "Updated by Wrong User",
			expectError: true,
		},
		{
			name:        "fail: invalid project ID format",
			projectID:   "invalid-uuid",
			userID:      userID,
			newName:     "Updated Name",
			expectError: true,
		},
		{
			name:        "fail: invalid user ID format",
			projectID:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			userID:      "invalid-uuid",
			newName:     "Updated Name",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updatedProject, err := storageInstance.UpdateProjectByUserID(ctx, tc.projectID, tc.userID, tc.newName)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.projectID, updatedProject.ID)
			assert.Equal(t, tc.expectedName, updatedProject.Name)
			assert.Equal(t, tc.userID, updatedProject.UserID)
		})
	}
}

func TestProjectStorageSQLc_DeleteProject(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	testCases := []struct {
		name        string
		projectID   string
		expectError bool
	}{
		{
			name:        "success: delete existing project",
			projectID:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			expectError: false,
		},
		{
			name:        "success: delete non-existent project (no error)",
			projectID:   "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "fail: invalid project ID format",
			projectID:   "invalid-uuid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database state
			truncateTables(t, db.Pool())
			seedTables(t, db.Pool())

			storageInstance := project.NewDefaultStorageSQLc(db)

			err := storageInstance.DeleteProject(ctx, tc.projectID)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify project is deleted if it was an existing project
			if tc.projectID == "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12" {
				_, err := storageInstance.GetProjectByID(ctx, tc.projectID)
				assert.Error(t, err) // Should not be found
			}
		})
	}
}

func TestProjectStorageSQLc_DeleteProjectByUserID(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

	testCases := []struct {
		name        string
		projectID   string
		userID      string
		expectError bool
	}{
		{
			name:        "success: delete project by correct user",
			projectID:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			userID:      userID,
			expectError: false,
		},
		{
			name:        "success: delete project by wrong user (no error but no deletion)",
			projectID:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			userID:      "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "fail: invalid project ID format",
			projectID:   "invalid-uuid",
			userID:      userID,
			expectError: true,
		},
		{
			name:        "fail: invalid user ID format",
			projectID:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			userID:      "invalid-uuid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database state
			truncateTables(t, db.Pool())
			seedTables(t, db.Pool())

			storageInstance := project.NewDefaultStorageSQLc(db)

			err := storageInstance.DeleteProjectByUserID(ctx, tc.projectID, tc.userID)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify deletion result based on user permissions
			project, err := storageInstance.GetProjectByID(ctx, tc.projectID)
			if tc.userID == userID {
				// Correct user - project should be deleted
				assert.Error(t, err) // Should not be found
			} else {
				// Wrong user - project should still exist
				assert.NoError(t, err)
				assert.NotNil(t, project)
			}
		})
	}
}

func TestProjectStorageSQLc_CountProjectsByUserID(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	truncateTables(t, db.Pool())
	seedTables(t, db.Pool())

	storageInstance := project.NewDefaultStorageSQLc(db)
	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

	testCases := []struct {
		name          string
		userID        string
		expectError   bool
		expectedCount int64
	}{
		{
			name:          "success: count projects for existing user",
			userID:        userID,
			expectError:   false,
			expectedCount: 1,
		},
		{
			name:          "success: count projects for non-existent user",
			userID:        "550e8400-e29b-41d4-a716-446655440000",
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:        "fail: invalid user ID format",
			userID:      "invalid-uuid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := storageInstance.CountProjectsByUserID(ctx, tc.userID)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedCount, count)
		})
	}
}

func TestProjectStorageSQLc_Integration(t *testing.T) {
	ctx := context.Background()
	db := SetupTestDatabase(t)
	defer db.Close()

	truncateTables(t, db.Pool())
	seedTables(t, db.Pool())

	storageInstance := project.NewDefaultStorageSQLc(db)
	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

	// Test complete workflow
	// 1. Create a project
	newProject := &project.Project{
		Name: "Integration Test Project",
	}
	createdProject, err := storageInstance.CreateProject(ctx, newProject, userID)
	require.NoError(t, err)
	assert.Equal(t, "Integration Test Project", createdProject.Name)

	// 2. Get projects by user ID
	projects, err := storageInstance.GetProjectsByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, projects, 2) // seed project + new project

	// 3. Count projects
	count, err := storageInstance.CountProjectsByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// 4. Update the project
	updatedProject, err := storageInstance.UpdateProjectByUserID(ctx, createdProject.ID, userID, "Updated Integration Project")
	require.NoError(t, err)
	assert.Equal(t, "Updated Integration Project", updatedProject.Name)

	// 5. Get the project by ID
	retrievedProject, err := storageInstance.GetProjectByID(ctx, createdProject.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Integration Project", retrievedProject.Name)

	// 6. Delete the project
	err = storageInstance.DeleteProjectByUserID(ctx, createdProject.ID, userID)
	require.NoError(t, err)

	// 7. Verify it's deleted
	_, err = storageInstance.GetProjectByID(ctx, createdProject.ID)
	assert.Error(t, err) // Should not be found
}

// Local test helpers to avoid import cycle
func truncateTables(t *testing.T, pool storage.PgxPool) {
	t.Helper()

	query := `
		TRUNCATE TABLE processed_events, invoices, subscriptions, images, jobs, projects, users, plans RESTART IDENTITY CASCADE
	`
	_, err := pool.Exec(context.Background(), query)
	require.NoError(t, err)
}

func seedTables(t *testing.T, pool storage.PgxPool) {
	t.Helper()

	// This path is relative to the tests/integration directory
	seedSQL, err := os.ReadFile("testdata/seed.sql")
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), string(seedSQL))
	require.NoError(t, err)
}
