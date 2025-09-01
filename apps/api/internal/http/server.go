// Package http provides the HTTP server and route handlers.
package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/services"
	"github.com/virtual-staging-ai/api/internal/storage"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	db           *storage.DB
	s3Service    storage.S3Service
	imageService *services.ImageService
	echo         *echo.Echo
}

// NewServer creates and configures a new Echo server.
func NewServer(db *storage.DB, s3Service storage.S3Service, imageService *services.ImageService) *Server {
	e := echo.New()
	s := &Server{db: db, s3Service: s3Service, imageService: imageService, echo: e}

	// Register routes
	g := e.Group("/api")
	g.POST("/v1/projects", s.createProjectHandler)
	g.GET("/v1/projects", s.getProjectsHandler)
	g.GET("/v1/projects/:id", s.getProjectByIDHandler)
	g.PUT("/v1/projects/:id", s.updateProjectHandler)
	g.DELETE("/v1/projects/:id", s.deleteProjectHandler)

	// Upload routes
	g.POST("/v1/uploads/presign", s.presignUploadHandler)

	// Image routes
	g.POST("/v1/images", s.createImageHandler)
	g.GET("/v1/images/:id", s.getImageHandler)
	g.DELETE("/v1/images/:id", s.deleteImageHandler)
	g.GET("/v1/projects/:project_id/images", s.getProjectImagesHandler)

	// SSE routes
	g.GET("/v1/events", s.eventsHandler)

	// Stripe routes
	g.POST("/v1/stripe/webhook", s.stripeWebhookHandler)

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
