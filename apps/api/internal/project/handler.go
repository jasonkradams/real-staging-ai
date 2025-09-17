package project

import (
	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out handler_mock.go . Handler

// Handler defines the HTTP-level contract for project operations.
// Implementations should be wired to Echo routes in the server.
type Handler interface {
	Create(c echo.Context) error
	List(c echo.Context) error
	GetByID(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
}
