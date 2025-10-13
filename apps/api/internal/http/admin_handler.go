package http

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"github.com/real-staging-ai/api/internal/auth"
	"github.com/real-staging-ai/api/internal/logging"
	"github.com/real-staging-ai/api/internal/settings"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/user"
)

// AdminHandler handles admin-related HTTP requests.
type AdminHandler struct {
	settingsService settings.Service
	db              storage.Database
	log             logging.Logger
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(settingsService settings.Service, db storage.Database, log logging.Logger) *AdminHandler {
	return &AdminHandler{
		settingsService: settingsService,
		db:              db,
		log:             log,
	}
}

// ListModels handles GET /admin/models - Lists all available AI models.
func (h *AdminHandler) ListModels(c echo.Context) error {
	ctx := c.Request().Context()

	models, err := h.settingsService.ListAvailableModels(ctx)
	if err != nil {
		h.log.Error(ctx, "failed to list models", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list models")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"models": models,
	})
}

// GetActiveModel handles GET /admin/models/active - Gets the currently active model.
func (h *AdminHandler) GetActiveModel(c echo.Context) error {
	ctx := c.Request().Context()

	modelID, err := h.settingsService.GetActiveModel(ctx)
	if err != nil {
		h.log.Error(ctx, "failed to get active model", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get active model")
	}

	// Get full model info
	models, err := h.settingsService.ListAvailableModels(ctx)
	if err != nil {
		h.log.Error(ctx, "failed to list models", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get model info")
	}

	var activeModel *settings.ModelInfo
	for _, model := range models {
		if model.ID == modelID {
			activeModel = &model
			break
		}
	}

	if activeModel == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Active model not found in registry")
	}

	return c.JSON(http.StatusOK, activeModel)
}

// UpdateActiveModel handles PUT /admin/models/active - Updates the active model.
func (h *AdminHandler) UpdateActiveModel(c echo.Context) error {
	ctx := c.Request().Context()

	var req settings.UpdateSettingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate required field
	if req.Value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "value is required")
	}

	// Get user UUID from Auth0 sub
	userUUID, err := h.resolveUserUUID(c)
	if err != nil {
		h.log.Error(ctx, "failed to resolve user", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{
			"message": "User not authenticated",
		})
	}

	err = h.settingsService.UpdateActiveModel(ctx, req.Value, userUUID)
	if err != nil {
		h.log.Error(ctx, "failed to update active model", "error", err, "model_id", req.Value)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	h.log.Info(ctx, "active model updated", "model_id", req.Value, "user_uuid", userUUID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":  "Active model updated successfully",
		"model_id": req.Value,
	})
}

// ListSettings handles GET /admin/settings - Lists all settings.
func (h *AdminHandler) ListSettings(c echo.Context) error {
	ctx := c.Request().Context()

	settings, err := h.settingsService.ListSettings(ctx)
	if err != nil {
		h.log.Error(ctx, "failed to list settings", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list settings")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"settings": settings,
	})
}

// GetSetting handles GET /admin/settings/:key - Gets a specific setting.
func (h *AdminHandler) GetSetting(c echo.Context) error {
	ctx := c.Request().Context()
	key := c.Param("key")

	setting, err := h.settingsService.GetSetting(ctx, key)
	if err != nil {
		h.log.Error(ctx, "failed to get setting", "error", err, "key", key)
		return echo.NewHTTPError(http.StatusNotFound, "Setting not found")
	}

	return c.JSON(http.StatusOK, setting)
}

// UpdateSetting handles PUT /admin/settings/:key - Updates a setting.
func (h *AdminHandler) UpdateSetting(c echo.Context) error {
	ctx := c.Request().Context()
	key := c.Param("key")

	var req settings.UpdateSettingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate required field
	if req.Value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "value is required")
	}

	// Get user UUID from Auth0 sub
	userUUID, err := h.resolveUserUUID(c)
	if err != nil {
		h.log.Error(ctx, "failed to resolve user", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{
			"message": "User not authenticated",
		})
	}

	err = h.settingsService.UpdateSetting(ctx, key, req.Value, userUUID)
	if err != nil {
		h.log.Error(ctx, "failed to update setting", "error", err, "key", key)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	h.log.Info(ctx, "setting updated", "key", key, "user_uuid", userUUID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Setting updated successfully",
		"key":     key,
		"value":   req.Value,
	})
}

// resolveUserUUID looks up or creates a user based on Auth0 sub, returning the user's UUID.
func (h *AdminHandler) resolveUserUUID(c echo.Context) (string, error) {
	ctx := c.Request().Context()

	// Get Auth0 sub from JWT token
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		return "", err
	}

	// Look up user by Auth0 sub
	uRepo := user.NewDefaultRepository(h.db)
	existingUser, err := uRepo.GetByAuth0Sub(ctx, auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User doesn't exist, create them
			newUser, err := uRepo.Create(ctx, auth0Sub, "", "user")
			if err != nil {
				h.log.Error(ctx, "failed to create user", "error", err, "auth0_sub", auth0Sub)
				return "", err
			}
			return newUser.ID.String(), nil
		}
		h.log.Error(ctx, "failed to get user by auth0 sub", "error", err, "auth0_sub", auth0Sub)
		return "", err
	}

	return existingUser.ID.String(), nil
}
