//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpLib "github.com/virtual-staging-ai/api/internal/http"
	"github.com/virtual-staging-ai/api/internal/image"
	"github.com/virtual-staging-ai/api/internal/storage"
)

type ProjectResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Error            string                  `json:"error"`
	Message          string                  `json:"message"`
	ValidationErrors []ValidationErrorDetail `json:"validation_errors"`
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func TestCreateProject_Handlers(t *testing.T) {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name           string
		requestBody    any
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, response []byte)
	}{
		{
			name: "success: valid project creation",
			requestBody: map[string]any{
				"name": "Living Room Staging",
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, response []byte) {
				var project ProjectResponse
				err := json.Unmarshal(response, &project)
				require.NoError(t, err)
				assert.NotEmpty(t, project.ID)
				assert.Equal(t, "Living Room Staging", project.Name)
				assert.NotEmpty(t, project.UserID)
				assert.NotZero(t, project.CreatedAt)
				// Validate UUID format
				_, err = uuid.Parse(project.ID)
				assert.NoError(t, err)
				_, err = uuid.Parse(project.UserID)
				assert.NoError(t, err)
			},
		},
		{
			name: "success: project with maximum length name",
			requestBody: map[string]any{
				"name": strings.Repeat("A", 100),
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, response []byte) {
				var project ProjectResponse
				err := json.Unmarshal(response, &project)
				require.NoError(t, err)
				assert.Equal(t, strings.Repeat("A", 100), project.Name)
			},
		},
		{
			name:           "fail: missing name field",
			requestBody:    map[string]any{},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: empty name",
			requestBody: map[string]any{
				"name": "",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: name too long",
			requestBody: map[string]any{
				"name": strings.Repeat("A", 101),
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: name with only whitespace",
			requestBody: map[string]any{
				"name": "   ",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name:           "fail: malformed JSON",
			requestBody:    `{"name":}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
		{
			name: "fail: invalid field type",
			requestBody: map[string]any{
				"name": 12345,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup clean database state
			TruncateAllTables(ctx, db.GetPool())
			SeedDatabase(ctx, db.GetPool())

			s3ServiceMock, err := storage.NewS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)
			imageServiceMock := &image.ServiceMock{}
			server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

			// Prepare request body
			var body []byte
			if str, ok := tc.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tc.requestBody)
				require.NoError(t, err)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			// TODO: Add Authorization header when auth middleware is implemented
			rec := httptest.NewRecorder()

			// Execute request
			server.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Assert error response if expected
			if tc.expectedError != "" {
				var errResp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedError, errResp.Error)
				assert.NotEmpty(t, errResp.Message)
			}

			// Run custom validation if provided
			if tc.validate != nil {
				tc.validate(t, rec.Body.Bytes())
			}
		})
	}
}

func TestGetProjects_Handlers(t *testing.T) {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name           string
		setupData      func(t *testing.T, db *storage.DB)
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, response []byte)
	}{
		{
			name: "success: get projects with data",
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				SeedDatabase(ctx, db.GetPool())
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var listResp ProjectListResponse
				err := json.Unmarshal(response, &listResp)
				require.NoError(t, err)
				assert.Len(t, listResp.Projects, 1)
				assert.Equal(t, "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", listResp.Projects[0].ID)
				assert.Equal(t, "Test Project 1", listResp.Projects[0].Name)
			},
		},
		{
			name: "success: empty project list",
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				// Only seed users, no projects
				_, err := db.GetPool().Exec(context.Background(),
					`INSERT INTO users (id, auth0_sub, role) VALUES
					('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'auth0|testuser', 'user')`)
				require.NoError(t, err)
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var listResp ProjectListResponse
				err := json.Unmarshal(response, &listResp)
				require.NoError(t, err)
				assert.Len(t, listResp.Projects, 0)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup data
			tc.setupData(t, db)

			s3ServiceMock, err := storage.NewS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)
			imageServiceMock := &image.ServiceMock{}
			server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
			// TODO: Add Authorization header when auth middleware is implemented
			rec := httptest.NewRecorder()

			// Execute request
			server.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Assert error response if expected
			if tc.expectedError != "" {
				var errResp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedError, errResp.Error)
			}

			// Run custom validation
			if tc.validate != nil {
				tc.validate(t, rec.Body.Bytes())
			}
		})
	}
}

func TestGetProjectByID_Handlers(t *testing.T) {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	// Setup data for all tests
	TruncateAllTables(ctx, db.GetPool())
	SeedDatabase(ctx, db.GetPool())

	testCases := []struct {
		name           string
		projectID      string
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, response []byte)
	}{
		{
			name:           "success: get existing project",
			projectID:      "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var project ProjectResponse
				err := json.Unmarshal(response, &project)
				require.NoError(t, err)
				assert.Equal(t, "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", project.ID)
				assert.Equal(t, "Test Project 1", project.Name)
				assert.Equal(t, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", project.UserID)
			},
		},
		{
			name:           "fail: project not found",
			projectID:      "550e8400-e29b-41d4-a716-446655440000",
			expectedStatus: http.StatusNotFound,
			expectedError:  "not_found",
		},
		{
			name:           "fail: invalid UUID format",
			projectID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
		{
			name:           "fail: empty project ID",
			projectID:      "",
			expectedStatus: http.StatusNotFound, // Echo router behavior
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s3ServiceMock, err := storage.NewS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)
			imageServiceMock := &image.ServiceMock{}
			server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

			// Create request
			url := fmt.Sprintf("/api/v1/projects/%s", tc.projectID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			// TODO: Add Authorization header when auth middleware is implemented
			rec := httptest.NewRecorder()

			// Execute request
			server.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Assert error response if expected
			if tc.expectedError != "" {
				var errResp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedError, errResp.Error)
			}

			// Run custom validation
			if tc.validate != nil {
				tc.validate(t, rec.Body.Bytes())
			}
		})
	}
}

func TestUpdateProject_Handlers(t *testing.T) {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name           string
		projectID      string
		requestBody    any
		setupData      func(t *testing.T, db *storage.DB)
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, response []byte)
	}{
		{
			name:      "success: update existing project",
			projectID: "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			requestBody: map[string]any{
				"name": "Updated Project Name",
			},
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				SeedDatabase(ctx, db.GetPool())
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var project ProjectResponse
				err := json.Unmarshal(response, &project)
				require.NoError(t, err)
				assert.Equal(t, "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", project.ID)
				assert.Equal(t, "Updated Project Name", project.Name)
			},
		},
		{
			name:      "fail: project not found",
			projectID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody: map[string]any{
				"name": "Updated Project Name",
			},
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				SeedDatabase(ctx, db.GetPool())
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "not_found",
		},
		{
			name:      "fail: invalid UUID format",
			projectID: "invalid-uuid",
			requestBody: map[string]any{
				"name": "Updated Project Name",
			},
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				SeedDatabase(ctx, db.GetPool())
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
		{
			name:      "fail: empty name",
			projectID: "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			requestBody: map[string]any{
				"name": "",
			},
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				SeedDatabase(ctx, db.GetPool())
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name:      "fail: name too long",
			projectID: "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			requestBody: map[string]any{
				"name": strings.Repeat("A", 101),
			},
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				SeedDatabase(ctx, db.GetPool())
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup data
			tc.setupData(t, db)

			s3ServiceMock, err := storage.NewS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)
			imageServiceMock := &image.ServiceMock{}
			server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

			// Prepare request body
			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			// Create request
			url := fmt.Sprintf("/api/v1/projects/%s", tc.projectID)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			// TODO: Add Authorization header when auth middleware is implemented
			rec := httptest.NewRecorder()

			// Execute request
			server.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Assert error response if expected
			if tc.expectedError != "" {
				var errResp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedError, errResp.Error)
			}

			// Run custom validation
			if tc.validate != nil {
				tc.validate(t, rec.Body.Bytes())
			}
		})
	}
}

func TestDeleteProject_Handlers(t *testing.T) {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name           string
		projectID      string
		setupData      func(t *testing.T, db *storage.DB)
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, db *storage.DB)
	}{
		{
			name:      "success: delete existing project",
			projectID: "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				SeedDatabase(ctx, db.GetPool())
			},
			expectedStatus: http.StatusNoContent,
			validate: func(t *testing.T, db *storage.DB) {
				// Verify project is deleted
				var count int
				err := db.GetPool().QueryRow(context.Background(),
					"SELECT COUNT(*) FROM projects WHERE id = $1",
					"b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12").Scan(&count)
				require.NoError(t, err)
				assert.Equal(t, 0, count)
			},
		},
		{
			name:      "fail: project not found",
			projectID: "550e8400-e29b-41d4-a716-446655440000",
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				SeedDatabase(ctx, db.GetPool())
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "not_found",
		},
		{
			name:      "fail: invalid UUID format",
			projectID: "invalid-uuid",
			setupData: func(t *testing.T, db *storage.DB) {
				TruncateAllTables(ctx, db.GetPool())
				SeedDatabase(ctx, db.GetPool())
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup data
			tc.setupData(t, db)

			s3ServiceMock, err := storage.NewS3Service(context.Background(), "test-bucket")
			require.NoError(t, err)
			imageServiceMock := &image.ServiceMock{}
			server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

			// Create request
			url := fmt.Sprintf("/api/v1/projects/%s", tc.projectID)
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			// TODO: Add Authorization header when auth middleware is implemented
			rec := httptest.NewRecorder()

			// Execute request
			server.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Assert error response if expected
			if tc.expectedError != "" {
				var errResp ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedError, errResp.Error)
			}

			// Run custom validation
			if tc.validate != nil {
				tc.validate(t, db)
			}
		})
	}
}

// TestProjectCRUDFlow tests the complete CRUD workflow
func TestProjectCRUDFlow_Handlers(t *testing.T) {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	// Setup clean state
	TruncateAllTables(ctx, db.GetPool())
	SeedDatabase(ctx, db.GetPool())

	s3ServiceMock, err := storage.NewS3Service(context.Background(), "test-bucket")
	require.NoError(t, err)
	imageServiceMock := &image.ServiceMock{}
	server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

	// Step 1: Create a project
	createBody := map[string]any{
		"name": "CRUD Test Project",
	}
	body, err := json.Marshal(createBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var createdProject ProjectResponse
	err = json.Unmarshal(rec.Body.Bytes(), &createdProject)
	require.NoError(t, err)
	assert.Equal(t, "CRUD Test Project", createdProject.Name)

	// Step 2: Get the project by ID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+createdProject.ID, nil)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var retrievedProject ProjectResponse
	err = json.Unmarshal(rec.Body.Bytes(), &retrievedProject)
	require.NoError(t, err)
	assert.Equal(t, createdProject.ID, retrievedProject.ID)
	assert.Equal(t, "CRUD Test Project", retrievedProject.Name)

	// Step 3: Update the project
	updateBody := map[string]any{
		"name": "Updated CRUD Test Project",
	}
	body, err = json.Marshal(updateBody)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+createdProject.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var updatedProject ProjectResponse
	err = json.Unmarshal(rec.Body.Bytes(), &updatedProject)
	require.NoError(t, err)
	assert.Equal(t, createdProject.ID, updatedProject.ID)
	assert.Equal(t, "Updated CRUD Test Project", updatedProject.Name)

	// Step 4: List projects (should include our updated project)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var projectList ProjectListResponse
	err = json.Unmarshal(rec.Body.Bytes(), &projectList)
	require.NoError(t, err)
	assert.Len(t, projectList.Projects, 2) // seed project + our created project

	// Find our project in the list
	var foundProject *ProjectResponse
	for _, p := range projectList.Projects {
		if p.ID == createdProject.ID {
			foundProject = &p
			break
		}
	}
	require.NotNil(t, foundProject)
	assert.Equal(t, "Updated CRUD Test Project", foundProject.Name)

	// Step 5: Delete the project
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+createdProject.ID, nil)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)

	// Step 6: Verify project is deleted
	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+createdProject.ID, nil)
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
