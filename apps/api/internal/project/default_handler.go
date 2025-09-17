package project

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"github.com/virtual-staging-ai/api/internal/auth"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/user"
)

// DefaultHandler provides Echo HTTP handlers for project operations.
// It lives in the project package to follow the “handlers in their own package” pattern.
type DefaultHandler struct {
	db storage.Database
}

// NewDefaultHandler constructs a project HTTP handler backed by the provided DB.
func NewDefaultHandler(db storage.Database) *DefaultHandler {
	return &DefaultHandler{db: db}
}

// Ensure DefaultHandler implements Handler.
var _ Handler = (*DefaultHandler)(nil)

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ValidationErrorDetail represents a validation error for a specific field.
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents a validation error response.
type ValidationErrorResponse struct {
	Error            string                  `json:"error"`
	Message          string                  `json:"message"`
	ValidationErrors []ValidationErrorDetail `json:"validation_errors"`
}

// ProjectListResponse is the response envelope for list endpoints.
type ProjectListResponse struct {
	Projects []Project `json:"projects"`
}

// Create handles POST /api/v1/projects
func (h *DefaultHandler) Create(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
		})
	}

	if validationErrs := validateCreateProjectRequest(&req); len(validationErrs) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:            "validation_failed",
			Message:          "The provided data is invalid",
			ValidationErrors: validationErrs,
		})
	}

	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing JWT token",
		})
	}

	uRepo := user.NewDefaultRepository(h.db)
	u, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			u, err = uRepo.Create(c.Request().Context(), auth0Sub, "", "user")
			if err != nil {
				c.Logger().Errorf("Failed to create user: %v", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_server_error",
					Message: "Failed to create user",
				})
			}
		} else {
			c.Logger().Errorf("Failed to get user by auth0 sub: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to get user",
			})
		}
	}

	p := Project{Name: req.Name}

	repo := NewDefaultRepository(h.db)
	created, err := repo.CreateProject(c.Request().Context(), &p, u.ID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   fmt.Sprintf("internal_server_error > %v", err),
			Message: "Failed to create project",
		})
	}

	return c.JSON(http.StatusCreated, created)
}

// List handles GET /api/v1/projects
func (h *DefaultHandler) List(c echo.Context) error {
	if h.db == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing JWT token",
		})
	}
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing JWT token",
		})
	}

	uRepo := user.NewDefaultRepository(h.db)
	u, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			u, err = uRepo.Create(c.Request().Context(), auth0Sub, "", "user")
			if err != nil {
				c.Logger().Errorf("Failed to create user: %v", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_server_error",
					Message: "Failed to create user",
				})
			}
		} else {
			c.Logger().Errorf("Failed to get user by auth0 sub: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to get user",
			})
		}
	}

	repo := NewDefaultRepository(h.db)
	projects, err := repo.GetProjectsByUserID(c.Request().Context(), u.ID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve projects",
		})
	}

	return c.JSON(http.StatusOK, ProjectListResponse{Projects: projects})
}

// GetByID handles GET /api/v1/projects/:id
func (h *DefaultHandler) GetByID(c echo.Context) error {
	projectID := c.Param("id")

	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing JWT token",
		})
	}

	uRepo := user.NewDefaultRepository(h.db)
	u, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			u, err = uRepo.Create(c.Request().Context(), auth0Sub, "", "user")
			if err != nil {
				c.Logger().Errorf("Failed to create user: %v", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_server_error",
					Message: "Failed to create user",
				})
			}
		} else {
			c.Logger().Errorf("Failed to get user by auth0 sub: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to get user",
			})
		}
	}

	repo := NewDefaultRepository(h.db)
	p, err := repo.GetProjectByIDAndUserID(c.Request().Context(), projectID, u.ID.String())
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

	return c.JSON(http.StatusOK, p)
}

// Update handles PUT /api/v1/projects/:id
func (h *DefaultHandler) Update(c echo.Context) error {
	projectID := c.Param("id")

	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
		})
	}

	if validationErrs := validateUpdateProjectRequest(&req); len(validationErrs) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:            "validation_failed",
			Message:          "The provided data is invalid",
			ValidationErrors: validationErrs,
		})
	}

	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing JWT token",
		})
	}

	uRepo := user.NewDefaultRepository(h.db)
	u, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			u, err = uRepo.Create(c.Request().Context(), auth0Sub, "", "user")
			if err != nil {
				c.Logger().Errorf("Failed to create user: %v", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_server_error",
					Message: "Failed to create user",
				})
			}
		} else {
			c.Logger().Errorf("Failed to get user by auth0 sub: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to get user",
			})
		}
	}

	repo := NewDefaultRepository(h.db)
	updated, err := repo.UpdateProjectByUserID(c.Request().Context(), projectID, u.ID.String(), req.Name)
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

	return c.JSON(http.StatusOK, updated)
}

// Delete handles DELETE /api/v1/projects/:id
func (h *DefaultHandler) Delete(c echo.Context) error {
	projectID := c.Param("id")

	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing JWT token",
		})
	}

	uRepo := user.NewDefaultRepository(h.db)
	u, err := uRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			u, err = uRepo.Create(c.Request().Context(), auth0Sub, "", "user")
			if err != nil {
				c.Logger().Errorf("Failed to create user: %v", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_server_error",
					Message: "Failed to create user",
				})
			}
		} else {
			c.Logger().Errorf("Failed to get user by auth0 sub: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to get user",
			})
		}
	}

	repo := NewDefaultRepository(h.db)
	if err := repo.DeleteProjectByUserID(c.Request().Context(), projectID, u.ID.String()); err != nil {
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

func validateCreateProjectRequest(req *CreateRequest) []ValidationErrorDetail {
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

func validateUpdateProjectRequest(req *UpdateRequest) []ValidationErrorDetail {
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
