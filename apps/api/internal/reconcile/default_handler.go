package reconcile

import (
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

// DefaultHandler handles HTTP requests for reconciliation operations.
type DefaultHandler struct {
	service Service
}

// NewDefaultHandler creates a new DefaultHandler.
func NewDefaultHandler(service Service) *DefaultHandler {
	return &DefaultHandler{service: service}
}

// ReconcileImagesRequest contains the request parameters for image reconciliation.
type ReconcileImagesRequest struct {
	ProjectID *string `json:"project_id" query:"project_id"`
	Status    *string `json:"status" query:"status"`
	Limit     int     `json:"limit" query:"limit"`
	Cursor    *string `json:"cursor" query:"cursor"`
	DryRun    bool    `json:"dry_run" query:"dry_run"`
}

// ReconcileImages handles POST /api/v1/admin/reconcile/images.
func (h *DefaultHandler) ReconcileImages(c echo.Context) error {
	// Feature flag check
	if os.Getenv("RECONCILE_ENABLED") != "1" {
		return c.JSON(http.StatusNotImplemented, map[string]string{
			"error": "reconciliation is not enabled",
		})
	}

	// TODO: Add role-gated auth check when admin roles are implemented
	// For now, we rely on the RECONCILE_ENABLED flag for access control

	var req ReconcileImagesRequest

	// Try binding JSON body first
	if c.Request().Header.Get(echo.HeaderContentType) == echo.MIMEApplicationJSON {
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request",
			})
		}
	}

	// Then bind query parameters (they can override JSON)
	if err := (&echo.DefaultBinder{}).BindQueryParams(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid query parameters",
		})
	}

	// Default limit
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	// Parse concurrency from env or query (query overrides env)
	concurrency := 5
	if envConc := os.Getenv("RECONCILE_CONCURRENCY"); envConc != "" {
		if parsed, err := strconv.Atoi(envConc); err == nil && parsed > 0 {
			concurrency = parsed
		}
	}
	if concStr := c.QueryParam("concurrency"); concStr != "" {
		if parsed, err := strconv.Atoi(concStr); err == nil && parsed > 0 {
			concurrency = parsed
		}
	}

	opts := ReconcileOptions{
		ProjectID:   req.ProjectID,
		Status:      req.Status,
		Limit:       req.Limit,
		Cursor:      req.Cursor,
		DryRun:      req.DryRun,
		Concurrency: concurrency,
	}

	result, err := h.service.ReconcileImages(c.Request().Context(), opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}
