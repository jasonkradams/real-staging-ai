package image

import (
	"net/http"
	"slices"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DefaultHandler contains the HTTP handlers for image operations.
type DefaultHandler struct {
	service Service
}

// NewDefaultHandler creates a new Handler instance.
func NewDefaultHandler(service Service) *DefaultHandler {
	return &DefaultHandler{
		service: service,
	}
}

// CreateImage handles POST /api/v1/images requests.
func (h *DefaultHandler) CreateImage(c echo.Context) error {
	var req CreateImageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
		})
	}

	// Validate request
	if validationErrs := h.validateCreateImageRequest(&req); len(validationErrs) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:            "validation_failed",
			Message:          "The provided data is invalid",
			ValidationErrors: validationErrs,
		})
	}

	// Create the image
	img, err := h.service.CreateImage(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to create image",
		})
	}

	return c.JSON(http.StatusCreated, img)
}

// GetImage handles GET /api/v1/images/{id} requests.
func (h *DefaultHandler) GetImage(c echo.Context) error {
	imageID := c.Param("id")
	if imageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Image ID is required",
		})
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid image ID format",
		})
	}

	img, err := h.service.GetImageByID(c.Request().Context(), imageID)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "no rows in result set" {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Image not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to get image",
		})
	}

	return c.JSON(http.StatusOK, img)
}

// GetProjectImages handles GET /api/v1/projects/{project_id}/images requests.
func (h *DefaultHandler) GetProjectImages(c echo.Context) error {
	projectID := c.Param("project_id")
	if projectID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Project ID is required",
		})
	}

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	images, err := h.service.GetImagesByProjectID(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to get images",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"images": images,
	})
}

// DeleteImage handles DELETE /api/v1/images/{id} requests.
func (h *DefaultHandler) DeleteImage(c echo.Context) error {
	imageID := c.Param("id")
	if imageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Image ID is required",
		})
	}

	// Validate UUID format
	if _, err := uuid.Parse(imageID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid image ID format",
		})
	}

	err := h.service.DeleteImage(c.Request().Context(), imageID)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "no rows in result set" {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Image not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to delete image",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// validateCreateImageRequest validates the create image request.
func (h *DefaultHandler) validateCreateImageRequest(req *CreateImageRequest) []ValidationErrorDetail {
	var errors []ValidationErrorDetail

	// Validate project ID
	if req.ProjectID == uuid.Nil {
		errors = append(errors, ValidationErrorDetail{
			Field:   "project_id",
			Message: "project_id is required",
		})
	}

	// Validate original URL
	if req.OriginalURL == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "original_url",
			Message: "original_url is required",
		})
	}

	// Validate room type if provided
	if req.RoomType != nil {
		validRoomTypes := []string{"living_room", "bedroom", "kitchen", "bathroom", "dining_room", "office", "entryway", "outdoor"}
		isValid := slices.Contains(validRoomTypes, *req.RoomType)
		if !isValid {
			errors = append(errors, ValidationErrorDetail{
				Field:   "room_type",
				Message: "room_type must be one of: living_room, bedroom, kitchen, bathroom, dining_room, office, entryway, outdoor",
			})
		}
	}

	// Validate style if provided
	if req.Style != nil {
		validStyles := []string{"modern", "contemporary", "traditional", "industrial", "scandinavian"}
		isValid := slices.Contains(validStyles, *req.Style)
		if !isValid {
			errors = append(errors, ValidationErrorDetail{
				Field:   "style",
				Message: "style must be one of: modern, contemporary, traditional, industrial, scandinavian",
			})
		}
	}

	// Validate seed if provided
	if req.Seed != nil {
		if *req.Seed < 1 || *req.Seed > 4294967295 {
			errors = append(errors, ValidationErrorDetail{
				Field:   "seed",
				Message: "seed must be between 1 and 4294967295",
			})
		}
	}

	return errors
}

// GetProjectCost handles GET /api/v1/projects/:project_id/cost requests.
func (h *DefaultHandler) GetProjectCost(c echo.Context) error {
	projectID := c.Param("project_id")

	// Validate project ID format
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid project ID format",
		})
	}

	// Get cost summary
	summary, err := h.service.GetProjectCostSummary(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve cost summary",
		})
	}

	return c.JSON(http.StatusOK, summary)
}
