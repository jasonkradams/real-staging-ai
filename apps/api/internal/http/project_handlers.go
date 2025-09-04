package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/auth"
	"github.com/virtual-staging-ai/api/internal/project"
	"github.com/virtual-staging-ai/api/internal/user"
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

	userRepo := user.NewUserRepository(s.db)
	user, err := userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User not found, create a new one
			user, err = userRepo.Create(c.Request().Context(), auth0Sub, "", "user")
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

	p := project.Project{
		Name: req.Name,
	}

	projectStorage := project.NewStorage(s.db)
	createdProject, err := projectStorage.CreateProject(c.Request().Context(), &p, user.ID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   fmt.Sprintf("internal_server_error > %v", err),
			Message: "Failed to create project",
		})
	}

	return c.JSON(http.StatusCreated, createdProject)
}

func (s *Server) getProjectsHandler(c echo.Context) error {
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing JWT token",
		})
	}

	userRepo := user.NewUserRepository(s.db)
	user, err := userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User not found, create a new one
			user, err = userRepo.Create(c.Request().Context(), auth0Sub, "", "user")
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

	projectStorage := project.NewStorage(s.db)
	projects, err := projectStorage.GetProjectsByUserID(c.Request().Context(), user.ID.String())
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

	userRepo := user.NewUserRepository(s.db)
	user, err := userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User not found, create a new one
			user, err = userRepo.Create(c.Request().Context(), auth0Sub, "", "user")
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

	projectStorage := project.NewStorage(s.db)
	project, err := projectStorage.GetProjectByIDAndUserID(c.Request().Context(), projectID, user.ID.String())
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

	userRepo := user.NewUserRepository(s.db)
	user, err := userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User not found, create a new one
			user, err = userRepo.Create(c.Request().Context(), auth0Sub, "", "user")
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

	projectStorage := project.NewStorage(s.db)
	updatedProject, err := projectStorage.UpdateProjectByUserID(c.Request().Context(), projectID, user.ID.String(), req.Name)
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

	userRepo := user.NewUserRepository(s.db)
	user, err := userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User not found, create a new one
			user, err = userRepo.Create(c.Request().Context(), auth0Sub, "", "user")
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

	projectStorage := project.NewStorage(s.db)
	err = projectStorage.DeleteProjectByUserID(c.Request().Context(), projectID, user.ID.String())
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
