// Package http provides the HTTP server and route handlers.
package http

import (
	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/project"
)

// NewServer creates and configures a new Echo server.
func NewServer() *echo.Echo {
	e := echo.New()

	// Register routes
	g := e.Group("/api")
	g.POST("/v1/projects", project.CreateProjectHandler)
	g.GET("/v1/projects", project.GetProjectsHandler)

	// Serve API documentation
	e.Static("/api/v1/docs", "../../web/api/v1")

	return e
}
