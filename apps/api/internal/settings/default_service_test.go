package settings

import (
	"context"
	"fmt"
	"testing"
)

func TestDefaultService_GetActiveModel(t *testing.T) {
	ctx := context.Background()

	t.Run("success: returns active model", func(t *testing.T) {
		repo := &RepositoryMock{
			GetByKeyFunc: func(ctx context.Context, key string) (*Setting, error) {
				if key != "active_model" {
					t.Errorf("expected key 'active_model', got %s", key)
				}
				return &Setting{
					Key:   "active_model",
					Value: "qwen/qwen-image-edit",
				}, nil
			},
		}

		service := NewDefaultService(repo)
		modelID, err := service.GetActiveModel(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if modelID != "qwen/qwen-image-edit" {
			t.Errorf("expected 'qwen/qwen-image-edit', got %s", modelID)
		}

		if len(repo.GetByKeyCalls()) != 1 {
			t.Errorf("expected 1 call to GetByKey, got %d", len(repo.GetByKeyCalls()))
		}
	})

	t.Run("fail: repository error", func(t *testing.T) {
		repo := &RepositoryMock{
			GetByKeyFunc: func(ctx context.Context, key string) (*Setting, error) {
				return nil, ErrSettingNotFound
			},
		}

		service := NewDefaultService(repo)
		_, err := service.GetActiveModel(ctx)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDefaultService_UpdateActiveModel(t *testing.T) {
	ctx := context.Background()

	t.Run("success: updates active model", func(t *testing.T) {
		repo := &RepositoryMock{
			GetByKeyFunc: func(ctx context.Context, key string) (*Setting, error) {
				return &Setting{
					Key:   "active_model",
					Value: "qwen/qwen-image-edit",
				}, nil
			},
			UpdateFunc: func(ctx context.Context, key, value, userID string) error {
				if key != "active_model" {
					t.Errorf("expected key 'active_model', got %s", key)
				}
				if value != "black-forest-labs/flux-kontext-max" {
					t.Errorf("expected 'black-forest-labs/flux-kontext-max', got %s", value)
				}
				if userID != "user123" {
					t.Errorf("expected 'user123', got %s", userID)
				}
				return nil
			},
		}

		service := NewDefaultService(repo)
		err := service.UpdateActiveModel(ctx, "black-forest-labs/flux-kontext-max", "user123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(repo.UpdateCalls()) != 1 {
			t.Errorf("expected 1 call to Update, got %d", len(repo.UpdateCalls()))
		}
	})

	t.Run("fail: invalid model ID", func(t *testing.T) {
		repo := &RepositoryMock{
			GetByKeyFunc: func(ctx context.Context, key string) (*Setting, error) {
				return &Setting{
					Key:   "active_model",
					Value: "qwen/qwen-image-edit",
				}, nil
			},
		}

		service := NewDefaultService(repo)
		err := service.UpdateActiveModel(ctx, "invalid/model", "user123")

		if err == nil {
			t.Fatal("expected error for invalid model ID")
		}

		if len(repo.UpdateCalls()) != 0 {
			t.Errorf("expected 0 calls to Update, got %d", len(repo.UpdateCalls()))
		}
	})
}

func TestDefaultService_ListAvailableModels(t *testing.T) {
	ctx := context.Background()

	t.Run("success: lists all models with active flag", func(t *testing.T) {
		repo := &RepositoryMock{
			GetByKeyFunc: func(ctx context.Context, key string) (*Setting, error) {
				return &Setting{
					Key:   "active_model",
					Value: "qwen/qwen-image-edit",
				}, nil
			},
		}

		service := NewDefaultService(repo)
		models, err := service.ListAvailableModels(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(models) != 2 {
			t.Fatalf("expected 2 models, got %d", len(models))
		}

		// Check Qwen model
		qwenFound := false
		fluxFound := false
		for _, model := range models {
			if model.ID == "qwen/qwen-image-edit" {
				qwenFound = true
				if !model.IsActive {
					t.Error("expected Qwen model to be active")
				}
			}
			if model.ID == "black-forest-labs/flux-kontext-max" {
				fluxFound = true
				if model.IsActive {
					t.Error("expected Flux model to not be active")
				}
			}
		}

		if !qwenFound {
			t.Error("expected Qwen model in list")
		}
		if !fluxFound {
			t.Error("expected Flux model in list")
		}
	})
}

var ErrSettingNotFound = fmt.Errorf("setting not found")

func TestDefaultService_GetSetting(t *testing.T) {
	ctx := context.Background()

	t.Run("success: returns setting", func(t *testing.T) {
		repo := &RepositoryMock{
			GetByKeyFunc: func(ctx context.Context, key string) (*Setting, error) {
				return &Setting{
					Key:   key,
					Value: "test-value",
				}, nil
			},
		}

		service := NewDefaultService(repo)
		setting, err := service.GetSetting(ctx, "test-key")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if setting.Key != "test-key" {
			t.Errorf("expected key 'test-key', got %s", setting.Key)
		}
	})
}

func TestDefaultService_UpdateSetting(t *testing.T) {
	ctx := context.Background()

	t.Run("success: updates setting", func(t *testing.T) {
		repo := &RepositoryMock{
			UpdateFunc: func(ctx context.Context, key, value, userID string) error {
				return nil
			},
		}

		service := NewDefaultService(repo)
		err := service.UpdateSetting(ctx, "test-key", "test-value", "user123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(repo.UpdateCalls()) != 1 {
			t.Errorf("expected 1 call to Update, got %d", len(repo.UpdateCalls()))
		}
	})
}

func TestDefaultService_ListSettings(t *testing.T) {
	ctx := context.Background()

	t.Run("success: lists all settings", func(t *testing.T) {
		repo := &RepositoryMock{
			ListFunc: func(ctx context.Context) ([]Setting, error) {
				return []Setting{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				}, nil
			},
		}

		service := NewDefaultService(repo)
		settings, err := service.ListSettings(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(settings) != 2 {
			t.Errorf("expected 2 settings, got %d", len(settings))
		}
	})
}
