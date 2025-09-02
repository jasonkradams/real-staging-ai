package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/project"
)

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

type ProjectListResponse struct {
	Projects []project.Project `json:"projects"`
}

func (s *Server) createProjectHandler(c echo.Context) error {
	var req project.CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
		})
	}

	// Validate request
	if validationErrs := validateCreateProjectRequest(&req); len(validationErrs) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:            "validation_failed",
			Message:          "The provided data is invalid",
			ValidationErrors: validationErrs,
		})
	}

	// Create project entity
	p := project.Project{
		Name: req.Name,
	}

	projectStorage := project.NewStorage(s.db)
	// TODO: Get user ID from JWT token when auth middleware is implemented
	userID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	createdProject, err := projectStorage.CreateProject(c.Request().Context(), &p, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to create project",
		})
	}

	return c.JSON(http.StatusCreated, createdProject)
}

func (s *Server) getProjectsHandler(c echo.Context) error {
	projectStorage := project.NewStorage(s.db)
	projects, err := projectStorage.GetProjects(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve projects",
		})
	}

	response := ProjectListResponse{
		Projects: projects,
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) getProjectByIDHandler(c echo.Context) error {
	projectID := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	projectStorage := project.NewStorage(s.db)
	project, err := projectStorage.GetProjectByID(c.Request().Context(), projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Project not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve project",
		})
	}

	return c.JSON(http.StatusOK, project)
}

func (s *Server) updateProjectHandler(c echo.Context) error {
	projectID := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	var req project.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
		})
	}

	// Validate request
	if validationErrs := validateUpdateProjectRequest(&req); len(validationErrs) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:            "validation_failed",
			Message:          "The provided data is invalid",
			ValidationErrors: validationErrs,
		})
	}

	projectStorage := project.NewStorage(s.db)
	updatedProject, err := projectStorage.UpdateProject(c.Request().Context(), projectID, req.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Project not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to update project",
		})
	}

	return c.JSON(http.StatusOK, updatedProject)
}

func (s *Server) deleteProjectHandler(c echo.Context) error {
	projectID := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	projectStorage := project.NewStorage(s.db)
	err := projectStorage.DeleteProject(c.Request().Context(), projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Project not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to delete project",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// Validation helpers
func validateCreateProjectRequest(req *project.CreateRequest) []ValidationErrorDetail {
	var errors []ValidationErrorDetail

	name := strings.TrimSpace(req.Name)
	if name == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "name",
			Message: "name is required",
		})
	} else if len(name) > 100 {
		errors = append(errors, ValidationErrorDetail{
			Field:   "name",
			Message: "name must be 100 characters or less",
		})
	}

	return errors
}

func validateUpdateProjectRequest(req *project.UpdateRequest) []ValidationErrorDetail {
	var errors []ValidationErrorDetail

	name := strings.TrimSpace(req.Name)
	if name == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "name",
			Message: "name is required",
		})
	} else if len(name) > 100 {
		errors = append(errors, ValidationErrorDetail{
			Field:   "name",
			Message: "name must be 100 characters or less",
		})
	}

	return errors
}
