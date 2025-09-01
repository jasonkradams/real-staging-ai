package main

import (
	"github.com/labstack/echo/v4"
	"github.com/virtual-staging-ai/api/internal/project"
)

// main is the entrypoint of the API server.
func main() {
	e := echo.New()

	e.POST("/v1/projects", project.CreateProjectHandler)

	e.Logger.Fatal(e.Start(":8080"))
}
