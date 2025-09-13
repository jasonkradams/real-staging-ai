// Package http provides the HTTP server and route handlers.
package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/virtual-staging-ai/api/internal/auth"
	"github.com/virtual-staging-ai/api/internal/image"
	"github.com/virtual-staging-ai/api/internal/storage"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	echo         *echo.Echo
	db           storage.Database
	s3Service    storage.S3Service
	imageHandler *ImageHandler
	imageService image.Service
	authConfig   *auth.Auth0Config
}

// NewServer creates and configures a new Echo server.
func NewServer(db storage.Database, s3Service storage.S3Service, imageService image.Service) *Server {
	e := echo.New()

	// Add OpenTelemetry middleware
	e.Use(otelecho.Middleware("virtual-staging-api"))

	// Add other middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize Auth0 config
	authConfig := auth.NewAuth0Config()

	imageHandler := NewImageHandler(imageService)

	s := &Server{db: db, s3Service: s3Service, imageService: imageService, echo: e, authConfig: authConfig, imageHandler: imageHandler}

	// Health check route
	e.GET("/health", s.healthCheck)

	// Register routes
	api := e.Group("/api/v1")

	// Public routes (no authentication required)
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
	protected.POST("/images", s.imageHandler.createImageHandler)
	protected.GET("/images/:id", s.imageHandler.getImageHandler)
	protected.DELETE("/images/:id", s.imageHandler.deleteImageHandler)
	protected.GET("/projects/:project_id/images", s.imageHandler.getProjectImagesHandler)

	// SSE routes
	protected.GET("/events", s.eventsHandler)

	// Serve API documentation
	e.Static("/api/v1/docs", "../../web/api/v1")

	return s
}

// NewTestServer creates a new Echo server for testing without Auth0 middleware.
func NewTestServer(db storage.Database, s3Service storage.S3Service, imageService image.Service) *Server {
	e := echo.New()

	// Add basic middleware (no Auth0 for testing)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	imageHandler := NewImageHandler(imageService)

	s := &Server{db: db, s3Service: s3Service, imageService: imageService, echo: e, authConfig: nil, imageHandler: imageHandler}

	// Health check route (same as main server)
	e.GET("/health", s.healthCheck)

	// Register routes without authentication
	api := e.Group("/api/v1")

	// All routes are public for testing
	api.POST("/stripe/webhook", s.stripeWebhookHandler)

	// Project routes (no auth required for testing)
	api.POST("/projects", withTestUser(s.createProjectHandler))
	api.GET("/projects", withTestUser(s.getProjectsHandler))
	api.GET("/projects/:id", withTestUser(s.getProjectByIDHandler))
	api.PUT("/projects/:id", withTestUser(s.updateProjectHandler))
	api.DELETE("/projects/:id", withTestUser(s.deleteProjectHandler))

	// Upload routes
	api.POST("/uploads/presign", s.presignUploadHandler)

	// Image routes
	api.POST("/images", s.imageHandler.createImageHandler)
	api.GET("/images/:id", s.imageHandler.getImageHandler)
	api.DELETE("/images/:id", s.imageHandler.deleteImageHandler)
	api.GET("/projects/:project_id/images", s.imageHandler.getProjectImagesHandler)

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

// withTestUser ensures an X-Test-User header is present for test-only servers.
// It defaults to the seeded test user to keep integration tests deterministic.
func withTestUser(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("X-Test-User") == "" {
			c.Request().Header.Set("X-Test-User", "auth0|testuser")
		}
		return h(c)
	}
}
