package model

import (
	"testing"
)

func TestNewModelRegistry(t *testing.T) {
	t.Run("success: creates registry with default models", func(t *testing.T) {
		registry := NewModelRegistry()

		if registry == nil {
			t.Fatal("expected registry to be non-nil")
		}

		// Verify Qwen model is registered
		if !registry.Exists(ModelQwenImageEdit) {
			t.Error("expected Qwen model to be registered")
		}

		// Verify Flux Kontext model is registered
		if !registry.Exists(ModelFluxKontextMax) {
			t.Error("expected Flux Kontext model to be registered")
		}
	})

	t.Run("success: registry has correct model count", func(t *testing.T) {
		registry := NewModelRegistry()

		models := registry.List()
		if len(models) != 2 {
			t.Errorf("expected 2 models to be registered, got %d", len(models))
		}
	})
}

func TestModelRegistry_Register(t *testing.T) {
	t.Run("success: registers a new model", func(t *testing.T) {
		registry := &ModelRegistry{
			models: make(map[ModelID]*ModelMetadata),
		}

		metadata := &ModelMetadata{
			ID:           ModelID("test/model"),
			Name:         "Test Model",
			Description:  "A test model",
			Version:      "v1",
			InputBuilder: NewQwenInputBuilder(),
		}

		registry.Register(metadata)

		if !registry.Exists(ModelID("test/model")) {
			t.Error("expected model to be registered")
		}
	})

	t.Run("success: overwrites existing model", func(t *testing.T) {
		registry := &ModelRegistry{
			models: make(map[ModelID]*ModelMetadata),
		}

		metadata1 := &ModelMetadata{
			ID:          ModelID("test/model"),
			Name:        "Test Model v1",
			Description: "Version 1",
		}

		metadata2 := &ModelMetadata{
			ID:          ModelID("test/model"),
			Name:        "Test Model v2",
			Description: "Version 2",
		}

		registry.Register(metadata1)
		registry.Register(metadata2)

		model, err := registry.Get(ModelID("test/model"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.Name != "Test Model v2" {
			t.Errorf("expected name to be 'Test Model v2', got %s", model.Name)
		}
	})
}

func TestModelRegistry_Get(t *testing.T) {
	t.Run("success: retrieves Qwen model", func(t *testing.T) {
		registry := NewModelRegistry()

		model, err := registry.Get(ModelQwenImageEdit)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.ID != ModelQwenImageEdit {
			t.Errorf("expected model ID to be %s, got %s", ModelQwenImageEdit, model.ID)
		}

		if model.Name == "" {
			t.Error("expected model to have a name")
		}

		if model.InputBuilder == nil {
			t.Error("expected model to have an input builder")
		}
	})

	t.Run("success: retrieves Flux Kontext model", func(t *testing.T) {
		registry := NewModelRegistry()

		model, err := registry.Get(ModelFluxKontextMax)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if model.ID != ModelFluxKontextMax {
			t.Errorf("expected model ID to be %s, got %s", ModelFluxKontextMax, model.ID)
		}

		if model.Name != "Flux Kontext Max" {
			t.Errorf("expected model name to be 'Flux Kontext Max', got %s", model.Name)
		}
	})

	t.Run("fail: model not found", func(t *testing.T) {
		registry := NewModelRegistry()

		_, err := registry.Get(ModelID("nonexistent/model"))
		if err == nil {
			t.Fatal("expected error for nonexistent model")
		}

		expectedMsg := "model not found: nonexistent/model"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})
}

func TestModelRegistry_List(t *testing.T) {
	t.Run("success: lists all registered models", func(t *testing.T) {
		registry := NewModelRegistry()

		models := registry.List()

		if len(models) < 2 {
			t.Errorf("expected at least 2 models to be registered, got %d", len(models))
		}

		// Verify both models are in the list
		foundQwen := false
		foundFlux := false
		for _, model := range models {
			if model.ID == ModelQwenImageEdit {
				foundQwen = true
			}
			if model.ID == ModelFluxKontextMax {
				foundFlux = true
			}
		}

		if !foundQwen {
			t.Error("expected Qwen model to be in the list")
		}
		if !foundFlux {
			t.Error("expected Flux Kontext model to be in the list")
		}
	})

	t.Run("success: returns empty list for empty registry", func(t *testing.T) {
		registry := &ModelRegistry{
			models: make(map[ModelID]*ModelMetadata),
		}

		models := registry.List()

		if len(models) != 0 {
			t.Errorf("expected empty list, got %d models", len(models))
		}
	})
}

func TestModelRegistry_Exists(t *testing.T) {
	t.Run("success: returns true for Qwen model", func(t *testing.T) {
		registry := NewModelRegistry()

		if !registry.Exists(ModelQwenImageEdit) {
			t.Error("expected Qwen model to exist")
		}
	})

	t.Run("success: returns true for Flux Kontext model", func(t *testing.T) {
		registry := NewModelRegistry()

		if !registry.Exists(ModelFluxKontextMax) {
			t.Error("expected Flux Kontext model to exist")
		}
	})

	t.Run("success: returns false for nonexistent model", func(t *testing.T) {
		registry := NewModelRegistry()

		if registry.Exists(ModelID("nonexistent/model")) {
			t.Error("expected nonexistent model to not exist")
		}
	})
}
