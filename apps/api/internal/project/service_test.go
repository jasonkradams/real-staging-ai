package project_test

// import (
// 	"context"
// 	"errors"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"go.uber.org/mock/gomock"

// 	"github.com/virtual-staging-ai/api/internal/project"
// 	"github.com/virtual-staging-ai/api/internal/services"
// 	"github.com/virtual-staging-ai/api/internal/storage"
// 	"github.com/virtual-staging-ai/api/internal/storage/mocks"
// )

// func TestProjectService_CreateProject(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockProjectRepo := project.RepositoryMock(ctrl)
// 	mockS3Service := mocks.NewMockS3Service(ctrl)

// 	projectService := services.NewService(mockProjectRepo, mockS3Service)

// 	testCases := []struct {
// 		name        string
// 		request     *project.CreateRequest
// 		setupMocks  func()
// 		expectError bool
// 		errorMsg    string
// 		validate    func(t *testing.T, result *project.Project)
// 	}{
// 		{
// 			name: "success: create project with valid data",
// 			request: &project.CreateRequest{
// 				Name:   "Test Project",
// 				UserID: "user123",
// 			},
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					CreateProject(gomock.Any(), gomock.Any(), "user123").
// 					DoAndReturn(func(ctx context.Context, p *project.Project, userID string) (*project.Project, error) {
// 						return &project.Project{
// 							ID:     "project-id-123",
// 							Name:   p.Name,
// 							UserID: userID,
// 						}, nil
// 					}).
// 					Times(1)
// 			},
// 			expectError: false,
// 			validate: func(t *testing.T, result *project.Project) {
// 				assert.Equal(t, "project-id-123", result.ID)
// 				assert.Equal(t, "Test Project", result.Name)
// 				assert.Equal(t, "user123", result.UserID)
// 			},
// 		},
// 		{
// 			name: "failure: empty project name",
// 			request: &project.CreateRequest{
// 				Name:   "",
// 				UserID: "user123",
// 			},
// 			setupMocks: func() {
// 				// No mock calls expected since validation fails early
// 			},
// 			expectError: true,
// 			errorMsg:    "project name is required",
// 		},
// 		{
// 			name: "failure: empty user ID",
// 			request: &project.CreateRequest{
// 				Name:   "Test Project",
// 				UserID: "",
// 			},
// 			setupMocks: func() {
// 				// No mock calls expected since validation fails early
// 			},
// 			expectError: true,
// 			errorMsg:    "user ID is required",
// 		},
// 		{
// 			name: "failure: repository error",
// 			request: &project.CreateRequest{
// 				Name:   "Test Project",
// 				UserID: "user123",
// 			},
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					CreateProject(gomock.Any(), gomock.Any(), "user123").
// 					Return(nil, errors.New("database connection failed")).
// 					Times(1)
// 			},
// 			expectError: true,
// 			errorMsg:    "failed to create project",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.setupMocks()

// 			result, err := projectService.CreateProject(context.Background(), tc.request)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.Contains(t, err.Error(), tc.errorMsg)
// 				assert.Nil(t, result)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, result)
// 				if tc.validate != nil {
// 					tc.validate(t, result)
// 				}
// 			}
// 		})
// 	}
// }

// func TestProjectService_CreateProjectWithUpload(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
// 	mockS3Service := mocks.NewMockS3Service(ctrl)

// 	projectService := services.NewService(mockProjectRepo, mockS3Service)

// 	testCases := []struct {
// 		name        string
// 		request     *project.CreateRequest
// 		filename    string
// 		contentType string
// 		fileSize    int64
// 		setupMocks  func()
// 		expectError bool
// 		errorMsg    string
// 		validate    func(t *testing.T, result *services.WithUploadURL)
// 	}{
// 		{
// 			name: "success: create project with upload URL",
// 			request: &project.CreateRequest{
// 				Name:   "Upload Test Project",
// 				UserID: "user123",
// 			},
// 			filename:    "test.jpg",
// 			contentType: "image/jpeg",
// 			fileSize:    1024000,
// 			setupMocks: func() {
// 				// Expect project creation
// 				mockProjectRepo.EXPECT().
// 					CreateProject(gomock.Any(), gomock.Any(), "user123").
// 					Return(&project.Project{
// 						ID:     "project-456",
// 						Name:   "Upload Test Project",
// 						UserID: "user123",
// 					}, nil).
// 					Times(1)

// 				// Expect upload URL generation
// 				mockS3Service.EXPECT().
// 					GeneratePresignedUploadURL(gomock.Any(), "user123", "test.jpg", "image/jpeg", int64(1024000)).
// 					Return(&storage.PresignedUploadResult{
// 						UploadURL: "https://s3.example.com/upload-url",
// 						FileKey:   "uploads/user123/test-uuid.jpg",
// 						ExpiresIn: 900,
// 					}, nil).
// 					Times(1)
// 			},
// 			expectError: false,
// 			validate: func(t *testing.T, result *services.WithUploadURL) {
// 				assert.Equal(t, "project-456", result.Project.ID)
// 				assert.Equal(t, "Upload Test Project", result.Project.Name)
// 				assert.Equal(t, "https://s3.example.com/upload-url", result.UploadURL)
// 				assert.Equal(t, "uploads/user123/test-uuid.jpg", result.FileKey)
// 			},
// 		},
// 		{
// 			name: "partial success: project created but upload URL fails",
// 			request: &project.CreateRequest{
// 				Name:   "Partial Test Project",
// 				UserID: "user123",
// 			},
// 			filename:    "test.jpg",
// 			contentType: "image/jpeg",
// 			fileSize:    1024000,
// 			setupMocks: func() {
// 				// Expect project creation to succeed
// 				mockProjectRepo.EXPECT().
// 					CreateProject(gomock.Any(), gomock.Any(), "user123").
// 					Return(&project.Project{
// 						ID:     "project-789",
// 						Name:   "Partial Test Project",
// 						UserID: "user123",
// 					}, nil).
// 					Times(1)

// 				// Expect upload URL generation to fail
// 				mockS3Service.EXPECT().
// 					GeneratePresignedUploadURL(gomock.Any(), "user123", "test.jpg", "image/jpeg", int64(1024000)).
// 					Return(nil, errors.New("AWS credentials not configured")).
// 					Times(1)
// 			},
// 			expectError: true,
// 			errorMsg:    "project created but failed to generate upload URL",
// 			validate: func(t *testing.T, result *services.WithUploadURL) {
// 				// Should still return the created project
// 				assert.NotNil(t, result.Project)
// 				assert.Equal(t, "project-789", result.Project.ID)
// 				assert.Empty(t, result.UploadURL)
// 				assert.Empty(t, result.FileKey)
// 			},
// 		},
// 		{
// 			name: "failure: project creation fails",
// 			request: &project.CreateRequest{
// 				Name:   "Failed Project",
// 				UserID: "user123",
// 			},
// 			filename:    "test.jpg",
// 			contentType: "image/jpeg",
// 			fileSize:    1024000,
// 			setupMocks: func() {
// 				// Expect project creation to fail
// 				mockProjectRepo.EXPECT().
// 					CreateProject(gomock.Any(), gomock.Any(), "user123").
// 					Return(nil, errors.New("database error")).
// 					Times(1)

// 				// S3 service should not be called since project creation failed
// 			},
// 			expectError: true,
// 			errorMsg:    "failed to create project",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.setupMocks()

// 			result, err := projectService.CreateProjectWithUpload(
// 				context.Background(),
// 				tc.request,
// 				tc.filename,
// 				tc.contentType,
// 				tc.fileSize,
// 			)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.Contains(t, err.Error(), tc.errorMsg)
// 				if tc.validate != nil {
// 					tc.validate(t, result)
// 				}
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, result)
// 				if tc.validate != nil {
// 					tc.validate(t, result)
// 				}
// 			}
// 		})
// 	}
// }

// func TestProjectService_GetProjectsByUser(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
// 	mockS3Service := mocks.NewMockS3Service(ctrl)

// 	projectService := services.NewProjectService(mockProjectRepo, mockS3Service)

// 	testCases := []struct {
// 		name        string
// 		userID      string
// 		setupMocks  func()
// 		expectError bool
// 		errorMsg    string
// 		validate    func(t *testing.T, result []project.Project)
// 	}{
// 		{
// 			name:   "success: get projects for user with projects",
// 			userID: "user123",
// 			setupMocks: func() {
// 				expectedProjects := []project.Project{
// 					{ID: "proj1", Name: "Project 1", UserID: "user123"},
// 					{ID: "proj2", Name: "Project 2", UserID: "user123"},
// 				}
// 				mockProjectRepo.EXPECT().
// 					GetProjectsByUserID(gomock.Any(), "user123").
// 					Return(expectedProjects, nil).
// 					Times(1)
// 			},
// 			expectError: false,
// 			validate: func(t *testing.T, result []project.Project) {
// 				assert.Len(t, result, 2)
// 				assert.Equal(t, "Project 1", result[0].Name)
// 				assert.Equal(t, "Project 2", result[1].Name)
// 			},
// 		},
// 		{
// 			name:   "success: get projects for user with no projects",
// 			userID: "user456",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					GetProjectsByUserID(gomock.Any(), "user456").
// 					Return([]project.Project{}, nil).
// 					Times(1)
// 			},
// 			expectError: false,
// 			validate: func(t *testing.T, result []project.Project) {
// 				assert.Len(t, result, 0)
// 			},
// 		},
// 		{
// 			name:   "failure: empty user ID",
// 			userID: "",
// 			setupMocks: func() {
// 				// No mock calls expected
// 			},
// 			expectError: true,
// 			errorMsg:    "user ID is required",
// 		},
// 		{
// 			name:   "failure: repository error",
// 			userID: "user123",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					GetProjectsByUserID(gomock.Any(), "user123").
// 					Return(nil, errors.New("database connection failed")).
// 					Times(1)
// 			},
// 			expectError: true,
// 			errorMsg:    "failed to get projects for user",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.setupMocks()

// 			result, err := projectService.GetProjectsByUser(context.Background(), tc.userID)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.Contains(t, err.Error(), tc.errorMsg)
// 				assert.Nil(t, result)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, result)
// 				if tc.validate != nil {
// 					tc.validate(t, result)
// 				}
// 			}
// 		})
// 	}
// }

// func TestProjectService_GetProjectByID(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
// 	mockS3Service := mocks.NewMockS3Service(ctrl)

// 	projectService := services.NewProjectService(mockProjectRepo, mockS3Service)

// 	testCases := []struct {
// 		name        string
// 		projectID   string
// 		userID      string
// 		setupMocks  func()
// 		expectError bool
// 		errorMsg    string
// 		validate    func(t *testing.T, result *project.Project)
// 	}{
// 		{
// 			name:      "success: get existing project",
// 			projectID: "proj123",
// 			userID:    "user123",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					GetProjectByIDAndUserID(gomock.Any(), "proj123", "user123").
// 					Return(&project.Project{
// 						ID:     "proj123",
// 						Name:   "Retrieved Project",
// 						UserID: "user123",
// 					}, nil).
// 					Times(1)
// 			},
// 			expectError: false,
// 			validate: func(t *testing.T, result *project.Project) {
// 				assert.Equal(t, "proj123", result.ID)
// 				assert.Equal(t, "Retrieved Project", result.Name)
// 				assert.Equal(t, "user123", result.UserID)
// 			},
// 		},
// 		{
// 			name:      "failure: empty project ID",
// 			projectID: "",
// 			userID:    "user123",
// 			setupMocks: func() {
// 				// No mock calls expected
// 			},
// 			expectError: true,
// 			errorMsg:    "project ID is required",
// 		},
// 		{
// 			name:      "failure: empty user ID",
// 			projectID: "proj123",
// 			userID:    "",
// 			setupMocks: func() {
// 				// No mock calls expected
// 			},
// 			expectError: true,
// 			errorMsg:    "user ID is required",
// 		},
// 		{
// 			name:      "failure: project not found",
// 			projectID: "proj456",
// 			userID:    "user123",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					GetProjectByIDAndUserID(gomock.Any(), "proj456", "user123").
// 					Return(nil, errors.New("project not found")).
// 					Times(1)
// 			},
// 			expectError: true,
// 			errorMsg:    "failed to get project",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.setupMocks()

// 			result, err := projectService.GetProjectByID(context.Background(), tc.projectID, tc.userID)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.Contains(t, err.Error(), tc.errorMsg)
// 				assert.Nil(t, result)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, result)
// 				if tc.validate != nil {
// 					tc.validate(t, result)
// 				}
// 			}
// 		})
// 	}
// }

// func TestProjectService_UpdateProject(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
// 	mockS3Service := mocks.NewMockS3Service(ctrl)

// 	projectService := services.NewProjectService(mockProjectRepo, mockS3Service)

// 	testCases := []struct {
// 		name        string
// 		projectID   string
// 		userID      string
// 		newName     string
// 		setupMocks  func()
// 		expectError bool
// 		errorMsg    string
// 		validate    func(t *testing.T, result *project.Project)
// 	}{
// 		{
// 			name:      "success: update project name",
// 			projectID: "proj123",
// 			userID:    "user123",
// 			newName:   "Updated Project Name",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					UpdateProjectByUserID(gomock.Any(), "proj123", "user123", "Updated Project Name").
// 					Return(&project.Project{
// 						ID:     "proj123",
// 						Name:   "Updated Project Name",
// 						UserID: "user123",
// 					}, nil).
// 					Times(1)
// 			},
// 			expectError: false,
// 			validate: func(t *testing.T, result *project.Project) {
// 				assert.Equal(t, "proj123", result.ID)
// 				assert.Equal(t, "Updated Project Name", result.Name)
// 				assert.Equal(t, "user123", result.UserID)
// 			},
// 		},
// 		{
// 			name:      "failure: empty project ID",
// 			projectID: "",
// 			userID:    "user123",
// 			newName:   "New Name",
// 			setupMocks: func() {
// 				// No mock calls expected
// 			},
// 			expectError: true,
// 			errorMsg:    "project ID is required",
// 		},
// 		{
// 			name:      "failure: empty user ID",
// 			projectID: "proj123",
// 			userID:    "",
// 			newName:   "New Name",
// 			setupMocks: func() {
// 				// No mock calls expected
// 			},
// 			expectError: true,
// 			errorMsg:    "user ID is required",
// 		},
// 		{
// 			name:      "failure: empty new name",
// 			projectID: "proj123",
// 			userID:    "user123",
// 			newName:   "",
// 			setupMocks: func() {
// 				// No mock calls expected
// 			},
// 			expectError: true,
// 			errorMsg:    "project name is required",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.setupMocks()

// 			result, err := projectService.UpdateProject(context.Background(), tc.projectID, tc.userID, tc.newName)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.Contains(t, err.Error(), tc.errorMsg)
// 				assert.Nil(t, result)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, result)
// 				if tc.validate != nil {
// 					tc.validate(t, result)
// 				}
// 			}
// 		})
// 	}
// }

// func TestProjectService_DeleteProject(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
// 	mockS3Service := mocks.NewMockS3Service(ctrl)

// 	projectService := services.NewProjectService(mockProjectRepo, mockS3Service)

// 	testCases := []struct {
// 		name        string
// 		projectID   string
// 		userID      string
// 		setupMocks  func()
// 		expectError bool
// 		errorMsg    string
// 	}{
// 		{
// 			name:      "success: delete project",
// 			projectID: "proj123",
// 			userID:    "user123",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					DeleteProjectByUserID(gomock.Any(), "proj123", "user123").
// 					Return(nil).
// 					Times(1)
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name:      "failure: empty project ID",
// 			projectID: "",
// 			userID:    "user123",
// 			setupMocks: func() {
// 				// No mock calls expected
// 			},
// 			expectError: true,
// 			errorMsg:    "project ID is required",
// 		},
// 		{
// 			name:      "failure: empty user ID",
// 			projectID: "proj123",
// 			userID:    "",
// 			setupMocks: func() {
// 				// No mock calls expected
// 			},
// 			expectError: true,
// 			errorMsg:    "user ID is required",
// 		},
// 		{
// 			name:      "failure: repository error",
// 			projectID: "proj123",
// 			userID:    "user123",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					DeleteProjectByUserID(gomock.Any(), "proj123", "user123").
// 					Return(errors.New("project not found")).
// 					Times(1)
// 			},
// 			expectError: true,
// 			errorMsg:    "failed to delete project",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.setupMocks()

// 			err := projectService.DeleteProject(context.Background(), tc.projectID, tc.userID)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.Contains(t, err.Error(), tc.errorMsg)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestProjectService_GetProjectStats(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
// 	mockS3Service := mocks.NewMockS3Service(ctrl)

// 	projectService := services.NewProjectService(mockProjectRepo, mockS3Service)

// 	testCases := []struct {
// 		name        string
// 		userID      string
// 		setupMocks  func()
// 		expectError bool
// 		errorMsg    string
// 		validate    func(t *testing.T, result *services.ProjectStats)
// 	}{
// 		{
// 			name:   "success: get project stats",
// 			userID: "user123",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					CountProjectsByUserID(gomock.Any(), "user123").
// 					Return(int64(5), nil).
// 					Times(1)
// 			},
// 			expectError: false,
// 			validate: func(t *testing.T, result *services.ProjectStats) {
// 				assert.Equal(t, int64(5), result.TotalProjects)
// 			},
// 		},
// 		{
// 			name:   "success: get stats for user with no projects",
// 			userID: "user456",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					CountProjectsByUserID(gomock.Any(), "user456").
// 					Return(int64(0), nil).
// 					Times(1)
// 			},
// 			expectError: false,
// 			validate: func(t *testing.T, result *services.ProjectStats) {
// 				assert.Equal(t, int64(0), result.TotalProjects)
// 			},
// 		},
// 		{
// 			name:   "failure: empty user ID",
// 			userID: "",
// 			setupMocks: func() {
// 				// No mock calls expected
// 			},
// 			expectError: true,
// 			errorMsg:    "user ID is required",
// 		},
// 		{
// 			name:   "failure: repository error",
// 			userID: "user123",
// 			setupMocks: func() {
// 				mockProjectRepo.EXPECT().
// 					CountProjectsByUserID(gomock.Any(), "user123").
// 					Return(int64(0), errors.New("database error")).
// 					Times(1)
// 			},
// 			expectError: true,
// 			errorMsg:    "failed to count projects",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.setupMocks()

// 			result, err := projectService.GetProjectStats(context.Background(), tc.userID)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.Contains(t, err.Error(), tc.errorMsg)
// 				assert.Nil(t, result)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, result)
// 				if tc.validate != nil {
// 					tc.validate(t, result)
// 				}
// 			}
// 		})
// 	}
// }

// // TestProjectService_ComplexScenarios tests more complex business logic scenarios
// func TestProjectService_ComplexScenarios(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
// 	mockS3Service := mocks.NewMockS3Service(ctrl)

// 	projectService := services.NewProjectService(mockProjectRepo, mockS3Service)

// 	t.Run("scenario: user workflow - create, get, update, delete", func(t *testing.T) {
// 		userID := "workflow-user"

// 		// Step 1: Create project with upload
// 		mockProjectRepo.EXPECT().
// 			CreateProject(gomock.Any(), gomock.Any(), userID).
// 			Return(&project.Project{ID: "new-project", Name: "Workflow Project", UserID: userID}, nil)

// 		mockS3Service.EXPECT().
// 			GeneratePresignedUploadURL(gomock.Any(), userID, "image.jpg", "image/jpeg", int64(1024)).
// 			Return(&storage.PresignedUploadResult{UploadURL: "https://upload.url"}, nil)

// 		result, err := projectService.CreateProjectWithUpload(
// 			context.Background(),
// 			&project.CreateRequest{Name: "Workflow Project", UserID: userID},
// 			"image.jpg",
// 			"image/jpeg",
// 			1024,
// 		)
// 		require.NoError(t, err)
// 		projectID := result.Project.ID

// 		// Step 2: Get the created project
// 		mockProjectRepo.EXPECT().
// 			GetProjectByIDAndUserID(gomock.Any(), projectID, userID).
// 			Return(&project.Project{ID: projectID, Name: "Workflow Project", UserID: userID}, nil)

// 		retrievedProject, err := projectService.GetProjectByID(context.Background(), projectID, userID)
// 		require.NoError(t, err)
// 		assert.Equal(t, "Workflow Project", retrievedProject.Name)

// 		// Step 3: Update the project
// 		mockProjectRepo.EXPECT().
// 			UpdateProjectByUserID(gomock.Any(), projectID, userID, "Updated Workflow Project").
// 			Return(&project.Project{ID: projectID, Name: "Updated Workflow Project", UserID: userID}, nil)

// 		updatedProject, err := projectService.UpdateProject(context.Background(), projectID, userID, "Updated Workflow Project")
// 		require.NoError(t, err)
// 		assert.Equal(t, "Updated Workflow Project", updatedProject.Name)

// 		// Step 4: Delete the project
// 		mockProjectRepo.EXPECT().
// 			DeleteProjectByUserID(gomock.Any(), projectID, userID).
// 			Return(nil)

// 		err = projectService.DeleteProject(context.Background(), projectID, userID)
// 		assert.NoError(t, err)
// 	})
// }
