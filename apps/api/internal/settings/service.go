package settings

import (
	"context"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out service_mock.go . Service

// Service defines the interface for settings business logic.
type Service interface {
	// GetActiveModel retrieves the currently active AI model ID.
	GetActiveModel(ctx context.Context) (string, error)

	// UpdateActiveModel updates the active AI model.
	UpdateActiveModel(ctx context.Context, modelID, userID string) error

	// ListAvailableModels returns all available AI models.
	ListAvailableModels(ctx context.Context) ([]ModelInfo, error)

	// GetSetting retrieves a setting by key.
	GetSetting(ctx context.Context, key string) (*Setting, error)

	// UpdateSetting updates a setting value.
	UpdateSetting(ctx context.Context, key, value, userID string) error

	// ListSettings retrieves all settings.
	ListSettings(ctx context.Context) ([]Setting, error)
}
