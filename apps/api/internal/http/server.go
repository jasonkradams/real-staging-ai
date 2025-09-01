// Package http provides the HTTP server and route handlers.
package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/virtual-staging-ai/api/internal/auth"
	"github.com/virtual-staging-ai/api/internal/services"
	"github.com/virtual-staging-ai/api/internal/storage"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	echo         *echo.Echo
	db           *storage.DB
	s3Service    storage.S3Service
	imageService *services.ImageService
	authConfig   *auth.Auth0Config
}

// NewServer creates and configures a new Echo server.
func NewServer(db *storage.DB, s3Service storage.S3Service, imageService *services.ImageService) *Server {
	e := echo.New()

	// Add OpenTelemetry middleware
	e.Use(otelecho.Middleware("virtual-staging-api"))

	// Add other middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize Auth0 config
	authConfig := auth.NewAuth0Config()

	s := &Server{db: db, s3Service: s3Service, imageService: imageService, echo: e, authConfig: authConfig}

	// Register routes
	api := e.Group("/api/v1")

	// Public routes (no authentication required)
	api.GET("/health", s.healthCheck)
	api.POST("/stripe/webhook", s.stripeWebhookHandler)

	// Protected routes (require JWT authentication)
	protected := api.Group("")
	protected.Use(auth.JWTMiddleware(s.authConfig))

	// Project routes
	protected.POST("/projects", s.createProjectHandler)
	protected.GET("/projects", s.getProjectsHandler)
	protected.GET("/projects/:id", s.getProjectByIDHandler)
	protected.PUT("/projects/:id", s.updateProjectHandler)
	protected.DELETE("/projects/:id", s.deleteProjectHandler)

	// Upload routes
	protected.POST("/uploads/presign", s.presignUploadHandler)

	// Image routes
	protected.POST("/images", s.createImageHandler)
	protected.GET("/images/:id", s.getImageHandler)
	protected.DELETE("/images/:id", s.deleteImageHandler)
	protected.GET("/projects/:project_id/images", s.getProjectImagesHandler)

	// SSE routes
	protected.GET("/events", s.eventsHandler)

	// Serve API documentation
	e.Static("/api/v1/docs", "../../web/api/v1")

	return s
}

// NewTestServer creates a new Echo server for testing without Auth0 middleware.
func NewTestServer(db *storage.DB, s3Service storage.S3Service, imageService *services.ImageService) *Server {
	e := echo.New()

	// Add basic middleware (no Auth0 for testing)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	s := &Server{db: db, s3Service: s3Service, imageService: imageService, echo: e, authConfig: nil}

	// Register routes without authentication
	api := e.Group("/api/v1")

	// All routes are public for testing
	api.GET("/health", s.healthCheck)
	api.POST("/stripe/webhook", s.stripeWebhookHandler)

	// Project routes (no auth required for testing)
	api.POST("/projects", s.createProjectHandler)
	api.GET("/projects", s.getProjectsHandler)
	api.GET("/projects/:id", s.getProjectByIDHandler)
	api.PUT("/projects/:id", s.updateProjectHandler)
	api.DELETE("/projects/:id", s.deleteProjectHandler)

	// Upload routes
	api.POST("/uploads/presign", s.presignUploadHandler)

	// Image routes
	api.POST("/images", s.createImageHandler)
	api.GET("/images/:id", s.getImageHandler)
	api.DELETE("/images/:id", s.deleteImageHandler)
	api.GET("/projects/:project_id/images", s.getProjectImagesHandler)

	// SSE routes
	api.GET("/events", s.eventsHandler)

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

// healthCheck handles GET /api/v1/health requests.
func (s *Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "virtual-staging-api",
	})
}
