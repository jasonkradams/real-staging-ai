package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	httpLib "github.com/virtual-staging-ai/api/internal/http"
	"github.com/virtual-staging-ai/api/internal/project"
)

func TestNewServer(t *testing.T) {
	s := httpLib.NewServer()
	assert.NotNil(t, s)
}

func TestCreateProjectRoute(t *testing.T) {
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
			// Setup
			server := httpLib.NewServer()
			req := httptest.NewRequest(http.MethodPost, "/v1/projects", strings.NewReader(tc.body))
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
