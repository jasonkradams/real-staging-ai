// Package project provides functionality for managing projects.
package project

import (
	"time"
)

// Project represents a user's project.
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name" validate:"required,min=1,max=100"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateProjectRequest represents the request payload for creating a project.
type CreateProjectRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateProjectRequest represents the request payload for updating a project.
type UpdateProjectRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
