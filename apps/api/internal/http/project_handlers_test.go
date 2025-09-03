//go:build integration

package http_test

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

// Test response types
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

// Test helper functions
type testSetup struct {
	db     *storage.DB
	server http.Handler
}

func setupTest(t *testing.T) *testSetup {
	t.Helper()

	db, err := storage.NewDB(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	storage.ResetDatabase(context.Background(), db.GetPool())
	storage.SeedDatabase(context.Background(), db.GetPool())

	s3ServiceMock, err := storage.NewS3Service(context.Background(), "test-bucket")
	require.NoError(t, err)

	imageServiceMock := &image.ServiceMock{}
	server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

	return &testSetup{
		db:     db,
		server: server,
	}
}

func setupTestWithCleanDB(t *testing.T) *testSetup {
	t.Helper()

	setup := setupTest(t)
	storage.TruncateAllTables(context.Background(), setup.db.GetPool())
	storage.SeedDatabase(context.Background(), setup.db.GetPool())

	return setup
}

func makeRequest(t *testing.T, server http.Handler, method, url string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody []byte
	var err error

	if body != nil {
		if str, ok := body.(string); ok {
			reqBody = []byte(str)
		} else {
			reqBody, err = json.Marshal(body)
			require.NoError(t, err)
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewReader(reqBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// TODO: Add Authorization header when auth middleware is implemented

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	return rec
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedError string) {
	t.Helper()

	if expectedError == "" {
		return
	}

	var errResp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, expectedError, errResp.Error)
	assert.NotEmpty(t, errResp.Message)
}

func assertValidUUID(t *testing.T, id string) {
	t.Helper()
	_, err := uuid.Parse(id)
	assert.NoError(t, err)
}

// Test cases
func TestCreateProject(t *testing.T) {
	testCases := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, response []byte)
	}{
		{
			name: "success: valid project creation",
			requestBody: map[string]interface{}{
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
				assertValidUUID(t, project.ID)
				assertValidUUID(t, project.UserID)
			},
		},
		{
			name: "success: project with maximum length name",
			requestBody: map[string]interface{}{
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
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: empty name",
			requestBody: map[string]interface{}{
				"name": "",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: name too long",
			requestBody: map[string]interface{}{
				"name": strings.Repeat("A", 101),
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name: "fail: name with only whitespace",
			requestBody: map[string]interface{}{
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
			requestBody: map[string]interface{}{
				"name": 12345,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use fresh setup for each test case to avoid interference
			testSetup := setupTestWithCleanDB(t)

			rec := makeRequest(t, testSetup.server, http.MethodPost, "/api/v1/projects", tc.requestBody)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assertErrorResponse(t, rec, tc.expectedError)

			if tc.validate != nil {
				tc.validate(t, rec.Body.Bytes())
			}
		})
	}
}

func TestGetProjects(t *testing.T) {
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
				storage.TruncateAllTables(context.Background(), db.GetPool())
				storage.SeedDatabase(context.Background(), db.GetPool())
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
				storage.TruncateAllTables(context.Background(), db.GetPool())
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
			setup := setupTest(t)
			tc.setupData(t, setup.db)

			rec := makeRequest(t, setup.server, http.MethodGet, "/api/v1/projects", nil)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assertErrorResponse(t, rec, tc.expectedError)

			if tc.validate != nil {
				tc.validate(t, rec.Body.Bytes())
			}
		})
	}
}

func TestGetProjectByID(t *testing.T) {
	setup := setupTest(t)

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
			url := fmt.Sprintf("/api/v1/projects/%s", tc.projectID)
			rec := makeRequest(t, setup.server, http.MethodGet, url, nil)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assertErrorResponse(t, rec, tc.expectedError)

			if tc.validate != nil {
				tc.validate(t, rec.Body.Bytes())
			}
		})
	}
}

func TestUpdateProject(t *testing.T) {
	testCases := []struct {
		name           string
		projectID      string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, response []byte)
	}{
		{
			name:      "success: update existing project",
			projectID: "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			requestBody: map[string]interface{}{
				"name": "Updated Project Name",
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
			requestBody: map[string]interface{}{
				"name": "Updated Project Name",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "not_found",
		},
		{
			name:      "fail: invalid UUID format",
			projectID: "invalid-uuid",
			requestBody: map[string]interface{}{
				"name": "Updated Project Name",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
		{
			name:      "fail: empty name",
			projectID: "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			requestBody: map[string]interface{}{
				"name": "",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
		{
			name:      "fail: name too long",
			projectID: "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			requestBody: map[string]interface{}{
				"name": strings.Repeat("A", 101),
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "validation_failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := setupTestWithCleanDB(t)

			url := fmt.Sprintf("/api/v1/projects/%s", tc.projectID)
			rec := makeRequest(t, setup.server, http.MethodPut, url, tc.requestBody)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assertErrorResponse(t, rec, tc.expectedError)

			if tc.validate != nil {
				tc.validate(t, rec.Body.Bytes())
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {
	testCases := []struct {
		name           string
		projectID      string
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, db *storage.DB)
	}{
		{
			name:           "success: delete existing project",
			projectID:      "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := setupTestWithCleanDB(t)

			url := fmt.Sprintf("/api/v1/projects/%s", tc.projectID)
			rec := makeRequest(t, setup.server, http.MethodDelete, url, nil)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assertErrorResponse(t, rec, tc.expectedError)

			if tc.validate != nil {
				tc.validate(t, setup.db)
			}
		})
	}
}

// TestProjectCRUDFlow tests the complete CRUD workflow
func TestProjectCRUDFlow(t *testing.T) {
	setup := setupTest(t)

	// Step 1: Create a project
	createBody := map[string]interface{}{
		"name": "CRUD Test Project",
	}
	rec := makeRequest(t, setup.server, http.MethodPost, "/api/v1/projects", createBody)
	require.Equal(t, http.StatusCreated, rec.Code)

	var createdProject ProjectResponse
	err := json.Unmarshal(rec.Body.Bytes(), &createdProject)
	require.NoError(t, err)
	assert.Equal(t, "CRUD Test Project", createdProject.Name)

	// Step 2: Get the project by ID
	rec = makeRequest(t, setup.server, http.MethodGet, "/api/v1/projects/"+createdProject.ID, nil)
	require.Equal(t, http.StatusOK, rec.Code)

	var retrievedProject ProjectResponse
	err = json.Unmarshal(rec.Body.Bytes(), &retrievedProject)
	require.NoError(t, err)
	assert.Equal(t, createdProject.ID, retrievedProject.ID)
	assert.Equal(t, "CRUD Test Project", retrievedProject.Name)

	// Step 3: Update the project
	updateBody := map[string]interface{}{
		"name": "Updated CRUD Test Project",
	}
	rec = makeRequest(t, setup.server, http.MethodPut, "/api/v1/projects/"+createdProject.ID, updateBody)
	require.Equal(t, http.StatusOK, rec.Code)

	var updatedProject ProjectResponse
	err = json.Unmarshal(rec.Body.Bytes(), &updatedProject)
	require.NoError(t, err)
	assert.Equal(t, createdProject.ID, updatedProject.ID)
	assert.Equal(t, "Updated CRUD Test Project", updatedProject.Name)

	// Step 4: List projects (should include our updated project)
	rec = makeRequest(t, setup.server, http.MethodGet, "/api/v1/projects", nil)
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
	rec = makeRequest(t, setup.server, http.MethodDelete, "/api/v1/projects/"+createdProject.ID, nil)
	require.Equal(t, http.StatusNoContent, rec.Code)

	// Step 6: Verify project is deleted
	rec = makeRequest(t, setup.server, http.MethodGet, "/api/v1/projects/"+createdProject.ID, nil)
	require.Equal(t, http.StatusNotFound, rec.Code)
}
