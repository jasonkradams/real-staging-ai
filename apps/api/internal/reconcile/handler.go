package reconcile

import (
	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out handler_mock.go . Handler

// Handler defines the HTTP-level contract for reconciliation operations.
// Implementations should be wired to Echo routes in the server.
type Handler interface {
	ReconcileImages(c echo.Context) error
}
