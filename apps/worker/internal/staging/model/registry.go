package model

import (
	"context"
	"fmt"

	"github.com/replicate/replicate-go"
)

// ModelID represents a unique identifier for a supported AI model.
type ModelID string

// Supported models
const (
	ModelQwenImageEdit  ModelID = "qwen/qwen-image-edit"
	ModelFluxKontextMax ModelID = "black-forest-labs/flux-kontext-max"
)

// ModelInputRequest contains the parameters needed to build model input.
type ModelInputRequest struct {
	ImageDataURL string
	Prompt       string
	Seed         *int64
	// Future: additional parameters for other models
}

// ModelInputBuilder defines the interface for building model-specific input parameters.
//
//go:generate go run github.com/matryer/moq@v0.5.3 -out model_mock.go . ModelInputBuilder
type ModelInputBuilder interface {
	// BuildInput creates the input parameters for the model.
	BuildInput(ctx context.Context, req *ModelInputRequest) (replicate.PredictionInput, error)

	// Validate checks if the request is valid for this model.
	Validate(req *ModelInputRequest) error
}

// ModelMetadata contains information about a registered model.
type ModelMetadata struct {
	ID           ModelID
	Name         string
	Description  string
	Version      string
	InputBuilder ModelInputBuilder
}

// ModelRegistry manages the available AI models and their configurations.
type ModelRegistry struct {
	models map[ModelID]*ModelMetadata
}

// NewModelRegistry creates a new registry with all supported models.
func NewModelRegistry() *ModelRegistry {
	registry := &ModelRegistry{
		models: make(map[ModelID]*ModelMetadata),
	}

	// Register Qwen Image Edit model
	registry.Register(&ModelMetadata{
		ID:           ModelQwenImageEdit,
		Name:         "Qwen Image Edit",
		Description:  "Fast image editing model optimized for virtual staging",
		Version:      "latest",
		InputBuilder: NewQwenInputBuilder(),
	})

	// Register Flux Kontext Max model
	registry.Register(&ModelMetadata{
		ID:           ModelFluxKontextMax,
		Name:         "Flux Kontext Max",
		Description:  "High-quality image generation and editing with advanced context understanding",
		Version:      "latest",
		InputBuilder: NewFluxKontextInputBuilder(),
	})

	return registry
}

// Register adds a model to the registry.
func (r *ModelRegistry) Register(metadata *ModelMetadata) {
	r.models[metadata.ID] = metadata
}

// Get retrieves a model's metadata by ID.
func (r *ModelRegistry) Get(id ModelID) (*ModelMetadata, error) {
	model, exists := r.models[id]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", id)
	}
	return model, nil
}

// List returns all registered models.
func (r *ModelRegistry) List() []*ModelMetadata {
	models := make([]*ModelMetadata, 0, len(r.models))
	for _, model := range r.models {
		models = append(models, model)
	}
	return models
}

// Exists checks if a model is registered.
func (r *ModelRegistry) Exists(id ModelID) bool {
	_, exists := r.models[id]
	return exists
}
