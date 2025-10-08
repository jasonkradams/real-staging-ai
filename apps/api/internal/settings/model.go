package settings

import "time"

// Setting represents a system configuration setting.
type Setting struct {
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	Description *string    `json:"description,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
	UpdatedBy   *string    `json:"updated_by,omitempty"`
}

// ModelInfo represents information about an available AI model.
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	IsActive    bool   `json:"is_active"`
}

// UpdateSettingRequest represents a request to update a setting.
type UpdateSettingRequest struct {
	Value string `json:"value" validate:"required"`
}
