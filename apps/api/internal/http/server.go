// Package http provides the HTTP server and route handlers.
package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/virtual-staging-ai/api/internal/auth"
	"github.com/virtual-staging-ai/api/internal/billing"
	"github.com/virtual-staging-ai/api/internal/image"
	"github.com/virtual-staging-ai/api/internal/logging"
	"github.com/virtual-staging-ai/api/internal/project"
	"github.com/virtual-staging-ai/api/internal/reconcile"
	"github.com/virtual-staging-ai/api/internal/settings"
	"github.com/virtual-staging-ai/api/internal/sse"
	"github.com/virtual-staging-ai/api/internal/storage"
	"github.com/virtual-staging-ai/api/internal/stripe"
	webdocs "github.com/virtual-staging-ai/api/web"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	echo      *echo.Echo
	db        storage.Database
	s3Service storage.S3Service

	imageService image.Service
	authConfig   *auth.Auth0Config
	pubsub       PubSub
}

// NewServer creates and configures a new Echo server.
func NewServer(db storage.Database, s3Service storage.S3Service, imageService image.Service, auth0Domain, auth0Audience string) *Server {
	e := echo.New()

	// Add OpenTelemetry middleware
	e.Use(otelecho.Middleware("virtual-staging-api"))

	// Add other middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods: []string{
			http.MethodGet, http.MethodHead, http.MethodPut,
			http.MethodPatch, http.MethodPost, http.MethodDelete,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization,
		},
	}))

	// Initialize Auth0 config
	authConfig := auth.NewAuth0Config(auth0Domain, auth0Audience)

	imgHandler := image.NewDefaultHandler(imageService)

	// Initialize Pub/Sub (Redis) if configured
	var ps PubSub
	if p, err := NewDefaultPubSubFromEnv(); err == nil {
		ps = p
	}

	s := &Server{db: db, s3Service: s3Service, imageService: imageService, echo: e, authConfig: authConfig, pubsub: ps}

	// Health check route
	e.GET("/health", s.healthCheck)

	// Register routes
	api := e.Group("/api/v1")

	// Public routes (no authentication required)
	api.POST("/stripe/webhook", func(c echo.Context) error {
		sh := stripe.NewDefaultHandler(s.db)
		return sh.Webhook(c)
	})

	// Protected routes (require JWT authentication)
	protected := api.Group("")
	protected.Use(auth.JWTMiddleware(s.authConfig))

	// Project routes
	ph := project.NewDefaultHandler(s.db)
	protected.POST("/projects", ph.Create)
	protected.GET("/projects", ph.List)
	protected.GET("/projects/:id", ph.GetByID)
	protected.DELETE("/projects/:id", ph.Delete)

	// Upload routes
	protected.POST("/uploads/presign", s.presignUploadHandler)

	// Image routes
	protected.POST("/images", imgHandler.CreateImage)
	protected.GET("/images/:id", imgHandler.GetImage)
	protected.GET("/images/:id/presign", s.presignImageDownloadHandler)
	protected.DELETE("/images/:id", imgHandler.DeleteImage)
	protected.GET("/projects/:project_id/images", imgHandler.GetProjectImages)

	// SSE routes
	protected.GET("/events", func(c echo.Context) error {
		cfg := sse.Config{
			SubscribeTimeout: 2000000000,
		}
		h, err := sse.NewDefaultHandlerFromEnv(cfg)
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "pubsub not configured"})
		}
		return h.Events(c)
	})

	// Billing routes
	bh := billing.NewDefaultHandler(s.db)
	protected.GET("/billing/subscriptions", bh.GetMySubscriptions)
	protected.GET("/billing/invoices", bh.GetMyInvoices)

	// Admin routes (feature-flagged)
	admin := protected.Group("/admin")
	reconcileSvc := reconcile.NewDefaultService(s.db, s.s3Service)
	reconcileHandler := reconcile.NewDefaultHandler(reconcileSvc)
	admin.POST("/reconcile/images", reconcileHandler.ReconcileImages)

	// Admin settings routes
	settingsRepo := settings.NewDefaultRepository(s.db.Pool())
	settingsService := settings.NewDefaultService(settingsRepo)
	adminHandler := NewAdminHandler(settingsService, logging.Default())
	admin.GET("/models", adminHandler.ListModels)
	admin.GET("/models/active", adminHandler.GetActiveModel)
	admin.PUT("/models/active", adminHandler.UpdateActiveModel)
	admin.GET("/settings", adminHandler.ListSettings)
	admin.GET("/settings/:key", adminHandler.GetSetting)
	admin.PUT("/settings/:key", adminHandler.UpdateSetting)

	// Serve API documentation (embedded)
	webdocs.RegisterRoutes(e)

	return s
}

// NewTestServer creates a new Echo server for testing without Auth0 middleware.
func NewTestServer(db storage.Database, s3Service storage.S3Service, imageService image.Service) *Server {
	e := echo.New()

	// Add basic middleware (no Auth0 for testing)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	imgHandler := image.NewDefaultHandler(imageService)

	s := &Server{db: db, s3Service: s3Service, imageService: imageService, echo: e, authConfig: nil}

	// Health check route (same as main server)
	e.GET("/health", s.healthCheck)

	// Register routes without authentication
	api := e.Group("/api/v1")

	// All routes are public for testing
	api.POST("/stripe/webhook", func(c echo.Context) error {
		sh := stripe.NewDefaultHandler(s.db)
		return sh.Webhook(c)
	})

	// Project routes (no auth required for testing)
	ph := project.NewDefaultHandler(s.db)
	api.POST("/projects", withTestUser(ph.Create))
	api.GET("/projects", withTestUser(ph.List))
	api.GET("/projects/:id", withTestUser(ph.GetByID))
	api.PUT("/projects/:id", withTestUser(ph.Update))
	api.DELETE("/projects/:id", withTestUser(ph.Delete))

	// Upload routes
	api.POST("/uploads/presign", s.presignUploadHandler)

	// Image routes
	api.POST("/images", imgHandler.CreateImage)
	api.GET("/images/:id", imgHandler.GetImage)
	api.GET("/images/:id/presign", s.presignImageDownloadHandler)
	api.DELETE("/images/:id", imgHandler.DeleteImage)
	api.GET("/projects/:project_id/images", imgHandler.GetProjectImages)

	// SSE routes
	api.GET("/events", func(c echo.Context) error {
		cfg := sse.Config{
			SubscribeTimeout: 2000000000,
		}
		h, err := sse.NewDefaultHandlerFromEnv(cfg)
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "pubsub not configured"})
		}
		return h.Events(c)
	})

	// Billing routes (public in test server)
	bh := billing.NewDefaultHandler(s.db)
	api.GET("/billing/subscriptions", withTestUser(bh.GetMySubscriptions))
	api.GET("/billing/invoices", withTestUser(bh.GetMyInvoices))

	// Admin routes (public in test server, feature-flagged)
	admin := api.Group("/admin")
	reconcileSvc := reconcile.NewDefaultService(s.db, s.s3Service)
	reconcileHandler := reconcile.NewDefaultHandler(reconcileSvc)
	admin.POST("/reconcile/images", reconcileHandler.ReconcileImages)

	// Admin settings routes (test server)
	settingsRepo := settings.NewDefaultRepository(s.db.Pool())
	settingsService := settings.NewDefaultService(settingsRepo)
	adminHandler := NewAdminHandler(settingsService, logging.Default())
	admin.GET("/models", withTestUser(adminHandler.ListModels))
	admin.GET("/models/active", withTestUser(adminHandler.GetActiveModel))
	admin.PUT("/models/active", withTestUser(adminHandler.UpdateActiveModel))
	admin.GET("/settings", withTestUser(adminHandler.ListSettings))
	admin.GET("/settings/:key", withTestUser(adminHandler.GetSetting))
	admin.PUT("/settings/:key", withTestUser(adminHandler.UpdateSetting))

	// Serve API documentation (embedded)
	webdocs.RegisterRoutes(e)

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
