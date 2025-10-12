//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpLib "github.com/real-staging-ai/api/internal/http"
	"github.com/real-staging-ai/api/internal/image"
	"github.com/real-staging-ai/api/internal/project"
	"github.com/stretchr/testify/assert"
)

func TestCreateProjectRoute_HTTP(t *testing.T) {
	// Setup
	db := SetupTestDatabase(t)
	defer db.Close()

	TruncateAllTables(context.Background(), db.Pool())
	SeedDatabase(context.Background(), db.Pool())

	s3ServiceMock := SetupTestS3Service(t, context.Background())
	imageServiceMock := &image.ServiceMock{}
	server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

	testCases := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{
			name:         "success: happy path",
			body:         `{"name":"Test Project"}`,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "fail: bad JSON",
			body:         `{"name":}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User", "auth0|testuser-create")
			rec := httptest.NewRecorder()

			// Run the request through the server
			server.ServeHTTP(rec, req)

			// Assertions
			assert.Equal(t, tc.expectedCode, rec.Code)

			if tc.expectedCode == http.StatusCreated {
				var p project.Project
				err := json.Unmarshal(rec.Body.Bytes(), &p)
				assert.NoError(t, err)
				assert.NotEmpty(t, p.ID)
				assert.Equal(t, "Test Project", p.Name)
			}
		})
	}
}

func TestGetProjectsRoute_HTTP(t *testing.T) {
	// Setup
	db := SetupTestDatabase(t)
	defer db.Close()

	TruncateAllTables(context.Background(), db.Pool())
	SeedDatabase(context.Background(), db.Pool())

	s3ServiceMock := SetupTestS3Service(t, context.Background())
	imageServiceMock := &image.ServiceMock{}
	server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set("X-Test-User", "auth0|testuser")
	rec := httptest.NewRecorder()

	// Run the request through the server
	server.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	type ProjectListResponse struct {
		Projects []project.Project `json:"projects"`
	}

	var response ProjectListResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(response.Projects), 1)
	found := false
	for _, p := range response.Projects {
		if p.ID == "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12" && p.Name == "Test Project 1" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected seeded project to be present")
}

func TestGetProjectByIDRoute_HTTP(t *testing.T) {
	// Setup
	db := SetupTestDatabase(t)
	defer db.Close()

	TruncateAllTables(context.Background(), db.Pool())
	SeedDatabase(context.Background(), db.Pool())

	s3ServiceMock := SetupTestS3Service(t, context.Background())
	imageServiceMock := &image.ServiceMock{}
	server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

	// Test case 1: Get an existing project
	t.Run("success: happy path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", nil)
		req.Header.Set("X-Test-User", "auth0|testuser")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var p project.Project
		err := json.Unmarshal(rec.Body.Bytes(), &p)
		assert.NoError(t, err)
		assert.Equal(t, "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", p.ID)
		assert.Equal(t, "Test Project 1", p.Name)
	})

	// Test case 2: Get a non-existing project
	t.Run("fail: not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a00", nil)
		req.Header.Set("X-Test-User", "auth0|testuser")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
