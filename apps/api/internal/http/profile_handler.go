package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/real-staging-ai/api/internal/auth"
	"github.com/real-staging-ai/api/internal/logging"
	"github.com/real-staging-ai/api/internal/user"
)

// ProfileHandler handles user profile HTTP requests.
type ProfileHandler struct {
	profileService user.ProfileService
	log            logging.Logger
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(profileService user.ProfileService, log logging.Logger) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		log:            log,
	}
}

// GetProfile handles GET /api/v1/user/profile - Gets the authenticated user's profile.
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	ctx := c.Request().Context()

	// Get Auth0 subject from JWT token
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		h.log.Error(ctx, "failed to get auth0 subject", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Get profile by Auth0 subject
	profile, err := h.profileService.GetProfileByAuth0Sub(ctx, auth0Sub)
	if err != nil {
		h.log.Error(ctx, "failed to get user profile", "error", err, "auth0_sub", auth0Sub)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile")
	}

	return c.JSON(http.StatusOK, profile)
}

// UpdateProfile handles PATCH /api/v1/user/profile - Updates the authenticated user's profile.
func (h *ProfileHandler) UpdateProfile(c echo.Context) error {
	ctx := c.Request().Context()

	// Get Auth0 subject from JWT token
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		h.log.Error(ctx, "failed to get auth0 subject", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Get current profile to obtain user ID
	currentProfile, err := h.profileService.GetProfileByAuth0Sub(ctx, auth0Sub)
	if err != nil {
		h.log.Error(ctx, "failed to get current profile", "error", err, "auth0_sub", auth0Sub)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve profile")
	}

	// Parse request body
	var req user.ProfileUpdateRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error(ctx, "failed to bind request", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Update profile
	updated, err := h.profileService.UpdateProfile(ctx, currentProfile.ID, &req)
	if err != nil {
		h.log.Error(ctx, "failed to update profile", "error", err, "user_id", currentProfile.ID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
	}

	return c.JSON(http.StatusOK, updated)
}
