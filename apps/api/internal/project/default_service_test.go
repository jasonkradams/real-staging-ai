package project_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtual-staging-ai/api/internal/project"
	"github.com/virtual-staging-ai/api/internal/storage"
)

func TestProjectService_CreateProject(t *testing.T) {
	testCases := []struct {
		name        string
		request     *project.CreateRequest
		setupMock   func(*project.RepositoryMock)
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, result *project.Project)
	}{
		{
			name: "success: create project with valid data",
			request: &project.CreateRequest{
				Name:   "Test Project",
				UserID: "user123",
			},
			setupMock: func(mock *project.RepositoryMock) {
				mock.CreateProjectFunc = func(ctx context.Context, p *project.Project, userID string) (*project.Project, error) {
					return &project.Project{
						ID:     "project-id-123",
						Name:   p.Name,
						UserID: userID,
					}, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *project.Project) {
				assert.Equal(t, "project-id-123", result.ID)
				assert.Equal(t, "Test Project", result.Name)
				assert.Equal(t, "user123", result.UserID)
			},
		},
		{
			name: "failure: empty project name",
			request: &project.CreateRequest{
				Name:   "",
				UserID: "user123",
			},
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed since validation fails early
			},
			expectError: true,
			errorMsg:    "project name is required",
		},
		{
			name: "failure: empty user ID",
			request: &project.CreateRequest{
				Name:   "Test Project",
				UserID: "",
			},
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed since validation fails early
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
		{
			name: "failure: repository error",
			request: &project.CreateRequest{
				Name:   "Test Project",
				UserID: "user123",
			},
			setupMock: func(mock *project.RepositoryMock) {
				mock.CreateProjectFunc = func(ctx context.Context, p *project.Project, userID string) (*project.Project, error) {
					return nil, errors.New("database connection failed")
				}
			},
			expectError: true,
			errorMsg:    "failed to create project",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projectRepositoryMock := &project.RepositoryMock{}
			tc.setupMock(projectRepositoryMock)

			s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)

			projectService := project.NewDefaultService(projectRepositoryMock, s3ServiceMock)

			result, err := projectService.CreateProject(context.Background(), tc.request)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}

func TestProjectService_CreateProjectWithUpload(t *testing.T) {
	testCases := []struct {
		name        string
		request     *project.CreateRequest
		filename    string
		contentType string
		fileSize    int64
		setupMock   func(*project.RepositoryMock, *storage.S3ServiceMock)
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, result *project.WithUploadURL)
	}{
		{
			name: "success: create project with upload URL",
			request: &project.CreateRequest{
				Name:   "Upload Test Project",
				UserID: "user123",
			},
			filename:    "test.jpg",
			contentType: "image/jpeg",
			fileSize:    1024000,
			setupMock: func(projectMock *project.RepositoryMock, s3Mock *storage.S3ServiceMock) {
				projectMock.CreateProjectFunc = func(ctx context.Context, p *project.Project, userID string) (*project.Project, error) {
					return &project.Project{
						ID:     "project-456",
						Name:   p.Name,
						UserID: userID,
					}, nil
				}
				s3Mock.GeneratePresignedUploadURLFunc = func(ctx context.Context, userID string, filename string, contentType string, fileSize int64) (*storage.PresignedUploadResult, error) {
					return &storage.PresignedUploadResult{
						UploadURL: "https://s3.example.com/upload-url",
						FileKey:   "uploads/user123/test-uuid.jpg",
						ExpiresIn: 900,
					}, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *project.WithUploadURL) {
				assert.Equal(t, "project-456", result.Project.ID)
				assert.Equal(t, "Upload Test Project", result.Project.Name)
				assert.Equal(t, "https://s3.example.com/upload-url", result.UploadURL)
				assert.Equal(t, "uploads/user123/test-uuid.jpg", result.FileKey)
			},
		},
		{
			name: "partial success: project created but upload URL fails",
			request: &project.CreateRequest{
				Name:   "Partial Test Project",
				UserID: "user123",
			},
			filename:    "test.jpg",
			contentType: "image/jpeg",
			fileSize:    1024000,
			setupMock: func(projectMock *project.RepositoryMock, s3Mock *storage.S3ServiceMock) {
				projectMock.CreateProjectFunc = func(ctx context.Context, p *project.Project, userID string) (*project.Project, error) {
					return &project.Project{
						ID:     "project-789",
						Name:   p.Name,
						UserID: userID,
					}, nil
				}
				s3Mock.GeneratePresignedUploadURLFunc = func(ctx context.Context, userID string, filename string, contentType string, fileSize int64) (*storage.PresignedUploadResult, error) {
					return nil, errors.New("AWS credentials not configured")
				}
			},
			expectError: true,
			errorMsg:    "project created but failed to generate upload URL",
			validate: func(t *testing.T, result *project.WithUploadURL) {
				// Should still return the created project
				assert.NotNil(t, result.Project)
				assert.Equal(t, "project-789", result.Project.ID)
				assert.Empty(t, result.UploadURL)
				assert.Empty(t, result.FileKey)
			},
		},
		{
			name: "failure: project creation fails",
			request: &project.CreateRequest{
				Name:   "Failed Project",
				UserID: "user123",
			},
			filename:    "test.jpg",
			contentType: "image/jpeg",
			fileSize:    1024000,
			setupMock: func(projectMock *project.RepositoryMock, s3Mock *storage.S3ServiceMock) {
				projectMock.CreateProjectFunc = func(ctx context.Context, p *project.Project, userID string) (*project.Project, error) {
					return nil, errors.New("database error")
				}
				// S3 service should not be called since project creation failed
			},
			expectError: true,
			errorMsg:    "failed to create project",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projectRepositoryMock := &project.RepositoryMock{}
			s3ServiceMock := &storage.S3ServiceMock{}
			tc.setupMock(projectRepositoryMock, s3ServiceMock)

			projectService := project.NewDefaultService(projectRepositoryMock, s3ServiceMock)

			result, err := projectService.CreateProjectWithUpload(
				context.Background(),
				tc.request,
				tc.filename,
				tc.contentType,
				tc.fileSize,
			)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}

func TestProjectService_GetProjectsByUser(t *testing.T) {
	testCases := []struct {
		name        string
		userID      string
		setupMock   func(*project.RepositoryMock)
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, result []project.Project)
	}{
		{
			name:   "success: get projects for user with projects",
			userID: "user123",
			setupMock: func(mock *project.RepositoryMock) {
				mock.GetProjectsByUserIDFunc = func(ctx context.Context, userID string) ([]project.Project, error) {
					return []project.Project{
						{ID: "proj1", Name: "Project 1", UserID: "user123"},
						{ID: "proj2", Name: "Project 2", UserID: "user123"},
					}, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result []project.Project) {
				assert.Len(t, result, 2)
				assert.Equal(t, "Project 1", result[0].Name)
				assert.Equal(t, "Project 2", result[1].Name)
			},
		},
		{
			name:   "success: get projects for user with no projects",
			userID: "user456",
			setupMock: func(mock *project.RepositoryMock) {
				mock.GetProjectsByUserIDFunc = func(ctx context.Context, userID string) ([]project.Project, error) {
					return []project.Project{}, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result []project.Project) {
				assert.Len(t, result, 0)
			},
		},
		{
			name:   "failure: empty user ID",
			userID: "",
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
		{
			name:   "failure: repository error",
			userID: "user123",
			setupMock: func(mock *project.RepositoryMock) {
				mock.GetProjectsByUserIDFunc = func(ctx context.Context, userID string) ([]project.Project, error) {
					return nil, errors.New("database connection failed")
				}
			},
			expectError: true,
			errorMsg:    "failed to get projects for user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projectRepositoryMock := &project.RepositoryMock{}
			tc.setupMock(projectRepositoryMock)

			s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)

			projectService := project.NewDefaultService(projectRepositoryMock, s3ServiceMock)

			result, err := projectService.GetProjectsByUser(context.Background(), tc.userID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}

func TestProjectService_GetProjectByID(t *testing.T) {
	testCases := []struct {
		name        string
		projectID   string
		userID      string
		setupMock   func(*project.RepositoryMock)
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, result *project.Project)
	}{
		{
			name:      "success: get existing project",
			projectID: "proj123",
			userID:    "user123",
			setupMock: func(mock *project.RepositoryMock) {
				mock.GetProjectByIDAndUserIDFunc = func(ctx context.Context, projectID string, userID string) (*project.Project, error) {
					return &project.Project{
						ID:     "proj123",
						Name:   "Retrieved Project",
						UserID: "user123",
					}, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *project.Project) {
				assert.Equal(t, "proj123", result.ID)
				assert.Equal(t, "Retrieved Project", result.Name)
				assert.Equal(t, "user123", result.UserID)
			},
		},
		{
			name:      "failure: empty project ID",
			projectID: "",
			userID:    "user123",
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed
			},
			expectError: true,
			errorMsg:    "project ID is required",
		},
		{
			name:      "failure: empty user ID",
			projectID: "proj123",
			userID:    "",
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
		{
			name:      "failure: project not found",
			projectID: "proj456",
			userID:    "user123",
			setupMock: func(mock *project.RepositoryMock) {
				mock.GetProjectByIDAndUserIDFunc = func(ctx context.Context, projectID string, userID string) (*project.Project, error) {
					return nil, errors.New("project not found")
				}
			},
			expectError: true,
			errorMsg:    "failed to get project",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projectRepositoryMock := &project.RepositoryMock{}
			tc.setupMock(projectRepositoryMock)

			s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)

			projectService := project.NewDefaultService(projectRepositoryMock, s3ServiceMock)

			result, err := projectService.GetProjectByID(context.Background(), tc.projectID, tc.userID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}

func TestProjectService_UpdateProject(t *testing.T) {
	testCases := []struct {
		name        string
		projectID   string
		userID      string
		newName     string
		setupMock   func(*project.RepositoryMock)
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, result *project.Project)
	}{
		{
			name:      "success: update project name",
			projectID: "proj123",
			userID:    "user123",
			newName:   "Updated Project Name",
			setupMock: func(mock *project.RepositoryMock) {
				mock.UpdateProjectByUserIDFunc = func(ctx context.Context, projectID string, userID string, newName string) (*project.Project, error) {
					return &project.Project{
						ID:     "proj123",
						Name:   "Updated Project Name",
						UserID: "user123",
					}, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *project.Project) {
				assert.Equal(t, "proj123", result.ID)
				assert.Equal(t, "Updated Project Name", result.Name)
				assert.Equal(t, "user123", result.UserID)
			},
		},
		{
			name:      "failure: empty project ID",
			projectID: "",
			userID:    "user123",
			newName:   "New Name",
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed
			},
			expectError: true,
			errorMsg:    "project ID is required",
		},
		{
			name:      "failure: empty user ID",
			projectID: "proj123",
			userID:    "",
			newName:   "New Name",
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
		{
			name:      "failure: empty new name",
			projectID: "proj123",
			userID:    "user123",
			newName:   "",
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed
			},
			expectError: true,
			errorMsg:    "project name is required",
		},
		{
			name:      "failure: repository error on update",
			projectID: "proj123",
			userID:    "user123",
			newName:   "New Name",
			setupMock: func(mock *project.RepositoryMock) {
				mock.UpdateProjectByUserIDFunc = func(ctx context.Context, projectID string, userID string, newName string) (*project.Project, error) {
					return nil, errors.New("database error")
				}
			},
			expectError: true,
			errorMsg:    "failed to update project",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projectRepositoryMock := &project.RepositoryMock{}
			tc.setupMock(projectRepositoryMock)

			s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)

			projectService := project.NewDefaultService(projectRepositoryMock, s3ServiceMock)

			result, err := projectService.UpdateProject(context.Background(), tc.projectID, tc.userID, tc.newName)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}

func TestProjectService_DeleteProject(t *testing.T) {
	testCases := []struct {
		name        string
		projectID   string
		userID      string
		setupMock   func(*project.RepositoryMock)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "success: delete project",
			projectID: "proj123",
			userID:    "user123",
			setupMock: func(mock *project.RepositoryMock) {
				mock.DeleteProjectByUserIDFunc = func(ctx context.Context, projectID string, userID string) error {
					return nil
				}
			},
			expectError: false,
		},
		{
			name:      "failure: empty project ID",
			projectID: "",
			userID:    "user123",
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed
			},
			expectError: true,
			errorMsg:    "project ID is required",
		},
		{
			name:      "failure: empty user ID",
			projectID: "proj123",
			userID:    "",
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
		{
			name:      "failure: repository error",
			projectID: "proj123",
			userID:    "user123",
			setupMock: func(mock *project.RepositoryMock) {
				mock.DeleteProjectByUserIDFunc = func(ctx context.Context, projectID string, userID string) error {
					return errors.New("project not found")
				}
			},
			expectError: true,
			errorMsg:    "failed to delete project",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projectRepositoryMock := &project.RepositoryMock{}
			tc.setupMock(projectRepositoryMock)

			s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)

			projectService := project.NewDefaultService(projectRepositoryMock, s3ServiceMock)

			err = projectService.DeleteProject(context.Background(), tc.projectID, tc.userID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProjectService_GetProjectStats(t *testing.T) {
	testCases := []struct {
		name        string
		userID      string
		setupMock   func(*project.RepositoryMock)
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, result *project.ProjectStats)
	}{
		{
			name:   "success: get project stats",
			userID: "user123",
			setupMock: func(mock *project.RepositoryMock) {
				mock.CountProjectsByUserIDFunc = func(ctx context.Context, userID string) (int64, error) {
					return 5, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *project.ProjectStats) {
				assert.Equal(t, int64(5), result.TotalProjects)
			},
		},
		{
			name:   "success: get stats for user with no projects",
			userID: "user456",
			setupMock: func(mock *project.RepositoryMock) {
				mock.CountProjectsByUserIDFunc = func(ctx context.Context, userID string) (int64, error) {
					return 0, nil
				}
			},
			expectError: false,
			validate: func(t *testing.T, result *project.ProjectStats) {
				assert.Equal(t, int64(0), result.TotalProjects)
			},
		},
		{
			name:   "failure: empty user ID",
			userID: "",
			setupMock: func(mock *project.RepositoryMock) {
				// No mock setup needed
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
		{
			name:   "failure: repository error",
			userID: "user123",
			setupMock: func(mock *project.RepositoryMock) {
				mock.CountProjectsByUserIDFunc = func(ctx context.Context, userID string) (int64, error) {
					return 0, errors.New("database error")
				}
			},
			expectError: true,
			errorMsg:    "failed to count projects",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projectRepositoryMock := &project.RepositoryMock{}
			tc.setupMock(projectRepositoryMock)

			s3ServiceMock, err := storage.NewDefaultS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)

			projectService := project.NewDefaultService(projectRepositoryMock, s3ServiceMock)

			result, err := projectService.GetProjectStats(context.Background(), tc.userID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}

// TestProjectService_ComplexScenarios tests more complex business logic scenarios
func TestProjectService_ComplexScenarios(t *testing.T) {
	t.Run("scenario: user workflow - create, get, update, delete", func(t *testing.T) {
		userID := "workflow-user"

		// Step 1: Create project with upload
		projectRepositoryMock := &project.RepositoryMock{}
		projectRepositoryMock.CreateProjectFunc = func(ctx context.Context, p *project.Project, userID string) (*project.Project, error) {
			return &project.Project{
				ID:     "new-project",
				Name:   p.Name,
				UserID: userID,
			}, nil
		}

		s3ServiceMock := &storage.S3ServiceMock{}
		s3ServiceMock.GeneratePresignedUploadURLFunc = func(ctx context.Context, userID string, filename string, contentType string, fileSize int64) (*storage.PresignedUploadResult, error) {
			return &storage.PresignedUploadResult{
				UploadURL: "https://upload.url",
			}, nil
		}

		projectService := project.NewDefaultService(projectRepositoryMock, s3ServiceMock)

		result, err := projectService.CreateProjectWithUpload(
			context.Background(),
			&project.CreateRequest{Name: "Workflow Project", UserID: userID},
			"image.jpg",
			"image/jpeg",
			1024,
		)
		require.NoError(t, err)
		projectID := result.Project.ID

		// Step 2: Get the created project
		projectRepositoryMock.GetProjectByIDAndUserIDFunc = func(ctx context.Context, projectID string, userID string) (*project.Project, error) {
			return &project.Project{
				ID:     projectID,
				Name:   "Workflow Project",
				UserID: userID,
			}, nil
		}

		retrievedProject, err := projectService.GetProjectByID(context.Background(), projectID, userID)
		require.NoError(t, err)
		assert.Equal(t, "Workflow Project", retrievedProject.Name)

		// Step 3: Update the project
		projectRepositoryMock.UpdateProjectByUserIDFunc = func(ctx context.Context, projectID string, userID string, newName string) (*project.Project, error) {
			return &project.Project{
				ID:     projectID,
				Name:   "Updated Workflow Project",
				UserID: userID,
			}, nil
		}

		updatedProject, err := projectService.UpdateProject(context.Background(), projectID, userID, "Updated Workflow Project")
		require.NoError(t, err)
		assert.Equal(t, "Updated Workflow Project", updatedProject.Name)

		// Step 4: Delete the project
		projectRepositoryMock.DeleteProjectByUserIDFunc = func(ctx context.Context, projectID string, userID string) error {
			return nil
		}

		err = projectService.DeleteProject(context.Background(), projectID, userID)
		assert.NoError(t, err)
	})
}
