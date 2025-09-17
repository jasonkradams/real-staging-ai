package image

import (
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
	CreateImage(c echo.Context) error
	GetImage(c echo.Context) error
	GetProjectImages(c echo.Context) error
	DeleteImage(c echo.Context) error
}
