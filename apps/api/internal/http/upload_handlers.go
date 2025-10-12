package http

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"github.com/real-staging-ai/api/internal/auth"
	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/user"
)

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ValidationErrorDetail represents a validation error for a specific field.
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents a validation error response.
type ValidationErrorResponse struct {
	Error            string                  `json:"error"`
	Message          string                  `json:"message"`
	ValidationErrors []ValidationErrorDetail `json:"validation_errors"`
}

type PresignUploadRequest struct {
	Filename    string `json:"filename" validate:"required,min=1,max=255"`
	ContentType string `json:"content_type" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,min=1,max=10485760"`
}

type PresignUploadResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresIn int64  `json:"expires_in"`
}

func (s *Server) presignUploadHandler(c echo.Context) error {
	var req PresignUploadRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request format",
		})
	}

	// Validate request
	if validationErrs := validatePresignUploadRequest(&req); len(validationErrs) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
			Error:            "validation_failed",
			Message:          "The provided data is invalid",
			ValidationErrors: validationErrs,
		})
	}

	// Get user ID from JWT token (or default in tests), ensure user exists
	auth0Sub, err := auth.GetUserIDOrDefault(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing JWT token",
		})
	}

	userRepo := user.NewDefaultRepository(s.db)
	u, err := userRepo.GetByAuth0Sub(c.Request().Context(), auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User not found, create a new one
			u, err = userRepo.Create(c.Request().Context(), auth0Sub, "", "user")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_server_error",
					Message: "Failed to create user",
				})
			}
		} else {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_server_error",
				Message: "Failed to get user",
			})
		}
	}

	// Use authenticated user's ID for S3 key scoping
	userID := u.ID.String()

	// Generate presigned upload URL using injected S3 service
	result, err := s.s3Service.GeneratePresignedUploadURL(
		c.Request().Context(),
		userID,
		req.Filename,
		req.ContentType,
		req.FileSize,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_server_error",
			Message: fmt.Sprintf("Failed to generate upload URL: %v", err),
		})
	}

	response := PresignUploadResponse{
		UploadURL: result.UploadURL,
		FileKey:   result.FileKey,
		ExpiresIn: result.ExpiresIn,
	}

	return c.JSON(http.StatusOK, response)
}

// Validation helpers for upload requests
func validatePresignUploadRequest(req *PresignUploadRequest) []ValidationErrorDetail {
	var errors []ValidationErrorDetail

	// Validate filename
	filename := strings.TrimSpace(req.Filename)
	switch {
	case filename == "":
		errors = append(errors, ValidationErrorDetail{
			Field:   "filename",
			Message: "filename is required",
		})
	case len(filename) > 255:
		errors = append(errors, ValidationErrorDetail{
			Field:   "filename",
			Message: "filename must be 255 characters or less",
		})
	case !storage.ValidateFilename(filename):
		errors = append(errors, ValidationErrorDetail{
			Field:   "filename",
			Message: "filename must have a valid image extension (.jpg, .jpeg, .png, .webp)",
		})
	}

	// Validate content type
	if req.ContentType == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "content_type",
			Message: "content_type is required",
		})
	} else if !storage.ValidateContentType(req.ContentType) {
		errors = append(errors, ValidationErrorDetail{
			Field:   "content_type",
			Message: "content_type must be image/jpeg, image/png, or image/webp",
		})
	}

	// Validate file size
	if req.FileSize <= 0 {
		errors = append(errors, ValidationErrorDetail{
			Field:   "file_size",
			Message: "file_size must be greater than 0",
		})
	} else if !storage.ValidateFileSize(req.FileSize) {
		errors = append(errors, ValidationErrorDetail{
			Field:   "file_size",
			Message: "file_size must be between 1 byte and 10MB",
		})
	}

	// Validate content type matches file extension
	if req.Filename != "" && req.ContentType != "" {
		ext := strings.ToLower(filepath.Ext(req.Filename))
		expectedContentType := getContentTypeFromExtension(ext)
		// If we have a valid image content type but invalid extension, or vice versa
		if (expectedContentType == "" && storage.ValidateContentType(req.ContentType)) ||
			(expectedContentType != "" && req.ContentType != expectedContentType) {
			errors = append(errors, ValidationErrorDetail{
				Field:   "content_type",
				Message: fmt.Sprintf("content_type %s doesn't match file extension %s", req.ContentType, ext),
			})
		}
	}

	return errors
}

func getContentTypeFromExtension(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}
