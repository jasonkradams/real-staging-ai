package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed api/v1/* api/v1/**
var fsDocs embed.FS

// RegisterRoutes mounts the embedded docs at /api/v1/docs.
func RegisterRoutes(e *echo.Echo) {
	sub, err := fs.Sub(fsDocs, "api/v1")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(sub))
	e.GET("/api/v1/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/api/v1/docs/")
	})
	e.GET("/api/v1/docs/*", echo.WrapHandler(http.StripPrefix("/api/v1/docs/", fileServer)))
}
