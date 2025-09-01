package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/image"
	"github.com/virtual-staging-ai/api/internal/services"
)

// ImageHandlers contains the HTTP handlers for image operations.
type ImageHandlers struct {
	imageService *services.ImageService
}

// NewImageHandlers creates a new ImageHandlers instance.
func NewImageHandlers(imageService *services.ImageService) *ImageHandlers {
	return &ImageHandlers{
		imageService: imageService,
	}
}

// createImageHandler handles POST /api/v1/images requests.
func (s *Server) createImageHandler(c echo.Context) error {
	var req image.CreateImageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
		})
	}

	// Validate request
	if validationErrs := validateCreateImageRequest(&req); len(validationErrs) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:            "validation_failed",
			Message:          "The provided data is invalid",
			ValidationErrors: validationErrs,
		})
	}

	// Create the image
	img, err := s.imageService.CreateImage(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to create image",
		})
	}

	return c.JSON(http.StatusCreated, img)
}

// getImageHandler handles GET /api/v1/images/{id} requests.
func (s *Server) getImageHandler(c echo.Context) error {
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

	img, err := s.imageService.GetImageByID(c.Request().Context(), imageID)
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

// getProjectImagesHandler handles GET /api/v1/projects/{project_id}/images requests.
func (s *Server) getProjectImagesHandler(c echo.Context) error {
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

	images, err := s.imageService.GetImagesByProjectID(c.Request().Context(), projectID)
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

// deleteImageHandler handles DELETE /api/v1/images/{id} requests.
func (s *Server) deleteImageHandler(c echo.Context) error {
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

	err := s.imageService.DeleteImage(c.Request().Context(), imageID)
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
func validateCreateImageRequest(req *image.CreateImageRequest) []ValidationErrorDetail {
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
		validRoomTypes := []string{"living_room", "bedroom", "kitchen", "bathroom", "dining_room", "office"}
		isValid := false
		for _, valid := range validRoomTypes {
			if *req.RoomType == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, ValidationErrorDetail{
				Field:   "room_type",
				Message: "room_type must be one of: living_room, bedroom, kitchen, bathroom, dining_room, office",
			})
		}
	}

	// Validate style if provided
	if req.Style != nil {
		validStyles := []string{"modern", "contemporary", "traditional", "industrial", "scandinavian"}
		isValid := false
		for _, valid := range validStyles {
			if *req.Style == valid {
				isValid = true
				break
			}
		}
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
