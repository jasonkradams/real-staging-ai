package image

import (
	"context"

	"github.com/labstack/echo/v4"
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

//go:generate go run github.com/matryer/moq@v0.5.3 -out handler_mock.go . Handler

type Handler interface {
	CreateImage(ctx context.Context, req *CreateImageRequest) (*Image, error)
	GetImage(ctx context.Context, imageID string) (*Image, error)
	GetProjectImages(c echo.Context) error
	DeleteImage(ctx context.Context, imageID string) error
	GetImagesForUser(ctx context.Context, userID string) ([]*Image, error)
}
