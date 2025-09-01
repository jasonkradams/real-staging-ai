package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/project"
	"github.com/virtual-staging-ai/api/internal/storage"
)

func (s *Server) createProjectHandler(c echo.Context) error {
	var p project.Project
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	projectStorage := storage.NewProjectStorage(s.db)
	createdProject, err := projectStorage.CreateProject(c.Request().Context(), &p)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, createdProject)
}

func (s *Server) getProjectsHandler(c echo.Context) error {
	projectStorage := storage.NewProjectStorage(s.db)
	projects, err := projectStorage.GetProjects(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, projects)
}
