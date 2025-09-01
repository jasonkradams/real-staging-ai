// Package http provides the HTTP server and route handlers.
package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/storage"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	db   *storage.DB
	echo *echo.Echo
}

// NewServer creates and configures a new Echo server.
func NewServer(db *storage.DB) *Server {
	e := echo.New()
	s := &Server{db: db, echo: e}

	// Register routes
	g := e.Group("/api")
	g.POST("/v1/projects", s.createProjectHandler)
	g.GET("/v1/projects", s.getProjectsHandler)

	// Serve API documentation
	e.Static("/api/v1/docs", "../../web/api/v1")

	return s
}

// Start starts the HTTP server.
func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.echo.ServeHTTP(w, r)
}
