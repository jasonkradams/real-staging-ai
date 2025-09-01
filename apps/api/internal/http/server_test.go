//go:build integration

package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	httpLib "github.com/virtual-staging-ai/api/internal/http"
	"github.com/virtual-staging-ai/api/internal/project"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/testutil"
)

func TestCreateProjectRoute(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	assert.NoError(t, err)
	defer db.Close()

	testutil.TruncateTables(t, db.GetPool())
	testutil.SeedTables(t, db.GetPool())

	mockS3Service := testutil.CreateMockS3Service(t)
	server := httpLib.NewServer(db, mockS3Service)

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

func TestGetProjectsRoute(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	assert.NoError(t, err)
	defer db.Close()

	testutil.TruncateTables(t, db.GetPool())
	testutil.SeedTables(t, db.GetPool())

	mockS3Service := testutil.CreateMockS3Service(t)
	server := httpLib.NewServer(db, mockS3Service)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec := httptest.NewRecorder()

	// Run the request through the server
	server.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	type ProjectListResponse struct {
		Projects []project.Project `json:"projects"`
	}

	var response ProjectListResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 1)
	assert.Equal(t, "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", response.Projects[0].ID)
	assert.Equal(t, "Test Project 1", response.Projects[0].Name)
}

func TestGetProjectByIDRoute(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	assert.NoError(t, err)
	defer db.Close()

	testutil.TruncateTables(t, db.GetPool())
	testutil.SeedTables(t, db.GetPool())

	mockS3Service := testutil.CreateMockS3Service(t)
	server := httpLib.NewServer(db, mockS3Service)

	// Test case 1: Get an existing project
	t.Run("success: happy path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var p project.Project
		err = json.Unmarshal(rec.Body.Bytes(), &p)
		assert.NoError(t, err)
		assert.Equal(t, "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", p.ID)
		assert.Equal(t, "Test Project 1", p.Name)
	})

	// Test case 2: Get a non-existing project
	t.Run("fail: not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a00", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
