package project_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-staging-ai/api/internal/project"
)

func TestCreateProject(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/v1/projects", strings.NewReader(`{"name":"Test Project"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, project.CreateProjectHandler(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// We can add more assertions here later, like checking the response body.
	}
}
