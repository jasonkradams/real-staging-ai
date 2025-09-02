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

// CreateRequest represents the input for creating a project.
type CreateRequest struct {
	Name   string `json:"name" validate:"required,min=1,max=100"`
	UserID string `json:"user_id" validate:"required"`
}

// UpdateRequest represents the request payload for updating a project.
type UpdateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
