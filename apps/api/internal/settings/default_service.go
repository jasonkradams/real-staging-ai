package settings

import (
	"context"
	"fmt"
)

// DefaultService implements Service.
type DefaultService struct {
	repo Repository
}

// Ensure DefaultService implements Service.
var _ Service = (*DefaultService)(nil)

// NewDefaultService creates a new DefaultService.
func NewDefaultService(repo Repository) *DefaultService {
	return &DefaultService{repo: repo}
}

// GetActiveModel retrieves the currently active AI model ID.
func (s *DefaultService) GetActiveModel(ctx context.Context) (string, error) {
	setting, err := s.repo.GetByKey(ctx, "active_model")
	if err != nil {
		return "", fmt.Errorf("failed to get active model: %w", err)
	}
	return setting.Value, nil
}

// UpdateActiveModel updates the active AI model.
func (s *DefaultService) UpdateActiveModel(ctx context.Context, modelID, userID string) error {
	// Validate model ID against available models
	models, err := s.ListAvailableModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	valid := false
	for _, model := range models {
		if model.ID == modelID {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid model ID: %s", modelID)
	}

	return s.repo.Update(ctx, "active_model", modelID, userID)
}

// ListAvailableModels returns all available AI models.
// This is hardcoded for now but could be made dynamic in the future.
func (s *DefaultService) ListAvailableModels(ctx context.Context) ([]ModelInfo, error) {
	// Get the current active model
	activeModelID, _ := s.GetActiveModel(ctx)

	models := []ModelInfo{
		{
			ID:          "qwen/qwen-image-edit",
			Name:        "Qwen Image Edit",
			Description: "Fast image editing model optimized for virtual staging. Requires input image.",
			Version:     "v1",
			IsActive:    activeModelID == "qwen/qwen-image-edit",
		},
		{
			ID:          "black-forest-labs/flux-kontext-max",
			Name:        "Flux Kontext Max",
			Description: "High-quality image generation and editing with advanced context understanding. Supports both text-to-image and image-to-image.",
			Version:     "v1",
			IsActive:    activeModelID == "black-forest-labs/flux-kontext-max",
		},
	}

	return models, nil
}

// GetSetting retrieves a setting by key.
func (s *DefaultService) GetSetting(ctx context.Context, key string) (*Setting, error) {
	return s.repo.GetByKey(ctx, key)
}

// UpdateSetting updates a setting value.
func (s *DefaultService) UpdateSetting(ctx context.Context, key, value, userID string) error {
	return s.repo.Update(ctx, key, value, userID)
}

// ListSettings retrieves all settings.
func (s *DefaultService) ListSettings(ctx context.Context) ([]Setting, error) {
	return s.repo.List(ctx)
}
