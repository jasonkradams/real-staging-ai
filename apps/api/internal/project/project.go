// Package project provides functionality for managing projects.
package project

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Project represents a user's project.
type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateProjectHandler handles the creation of a new project.
// It binds the request body to a Project struct and returns the created project.
func CreateProjectHandler(c echo.Context) error {
	var p Project
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = uuid.NewString() // Generate a new UUID for the project
	return c.JSON(http.StatusCreated, p)
}
