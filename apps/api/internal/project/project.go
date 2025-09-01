// Package project provides functionality for managing projects.
package project

import (
	"time"
)

// Project represents a user's project.
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
