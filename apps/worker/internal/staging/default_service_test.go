package staging

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/real-staging-ai/worker/internal/staging/model"
)

func TestNewDefaultService(t *testing.T) {
	ctx := context.Background()

	// Save original awsConfigLoader and restore after tests
	originalLoader := awsConfigLoader
	defer func() { awsConfigLoader = originalLoader }()

	// Mock AWS config loader
	awsConfigLoader = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{Region: "us-west-1"}, nil
	}

	t.Run("success: creates service with default model", func(t *testing.T) {
		cfg := &ServiceConfig{
			BucketName:     "test-bucket",
			ReplicateToken: "test-token",
			S3Endpoint:     "http://localhost:9000",
			S3Region:       "us-west-1",
			S3AccessKey:    "test-key",
			S3SecretKey:    "test-secret",
			S3UsePathStyle: true,
			AppEnv:         "dev",
		}

		service, err := NewDefaultService(ctx, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if service == nil {
			t.Fatal("expected service to be non-nil")
		}

		if service.modelID != model.ModelQwenImageEdit {
			t.Errorf("expected model ID to be %s, got %s", model.ModelQwenImageEdit, service.modelID)
		}

		if service.registry == nil {
			t.Error("expected registry to be non-nil")
		}
	})

	t.Run("success: creates service with specified model", func(t *testing.T) {
		cfg := &ServiceConfig{
			BucketName:     "test-bucket",
			ReplicateToken: "test-token",
			ModelID:        model.ModelQwenImageEdit,
			S3Endpoint:     "http://localhost:9000",
			S3Region:       "us-west-1",
			S3AccessKey:    "test-key",
			S3SecretKey:    "test-secret",
			S3UsePathStyle: true,
			AppEnv:         "dev",
		}

		service, err := NewDefaultService(ctx, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if service.modelID != model.ModelQwenImageEdit {
			t.Errorf("expected model ID to be %s, got %s", model.ModelQwenImageEdit, service.modelID)
		}
	})

	t.Run("success: creates service for test environment", func(t *testing.T) {
		cfg := &ServiceConfig{
			BucketName:     "test-bucket",
			ReplicateToken: "test-token",
			ModelID:        model.ModelQwenImageEdit,
			AppEnv:         "test",
		}

		service, err := NewDefaultService(ctx, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if service == nil {
			t.Fatal("expected service to be non-nil")
		}
	})

	t.Run("success: creates service for production environment", func(t *testing.T) {
		cfg := &ServiceConfig{
			BucketName:     "test-bucket",
			ReplicateToken: "test-token",
			ModelID:        model.ModelQwenImageEdit,
			AppEnv:         "prod",
		}

		service, err := NewDefaultService(ctx, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if service == nil {
			t.Fatal("expected service to be non-nil")
		}
	})

	t.Run("fail: missing bucket name", func(t *testing.T) {
		cfg := &ServiceConfig{
			ReplicateToken: "test-token",
		}

		_, err := NewDefaultService(ctx, cfg)
		if err == nil {
			t.Fatal("expected error for missing bucket name")
		}

		expectedMsg := "bucket name is required"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("fail: missing replicate token", func(t *testing.T) {
		cfg := &ServiceConfig{
			BucketName: "test-bucket",
		}

		_, err := NewDefaultService(ctx, cfg)
		if err == nil {
			t.Fatal("expected error for missing replicate token")
		}

		expectedMsg := "replicate API token is required"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("fail: unsupported model", func(t *testing.T) {
		cfg := &ServiceConfig{
			BucketName:     "test-bucket",
			ReplicateToken: "test-token",
			ModelID:        model.ModelID("unsupported/model"),
			S3Endpoint:     "http://localhost:9000",
			S3Region:       "us-west-1",
			S3AccessKey:    "test-key",
			S3SecretKey:    "test-secret",
		}

		_, err := NewDefaultService(ctx, cfg)
		if err == nil {
			t.Fatal("expected error for unsupported model")
		}

		expectedMsg := "unsupported model: unsupported/model"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("fail: AWS config load error", func(t *testing.T) {
		// Temporarily override the AWS config loader to return an error
		awsConfigLoader = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
			return aws.Config{}, errors.New("aws config error")
		}

		cfg := &ServiceConfig{
			BucketName:     "test-bucket",
			ReplicateToken: "test-token",
			ModelID:        model.ModelQwenImageEdit,
			S3Endpoint:     "http://localhost:9000",
			S3Region:       "us-west-1",
			S3AccessKey:    "test-key",
			S3SecretKey:    "test-secret",
		}

		_, err := NewDefaultService(ctx, cfg)
		if err == nil {
			t.Fatal("expected error for AWS config load failure")
		}

		// Restore the mock loader for other tests
		awsConfigLoader = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
			return aws.Config{Region: "us-west-1"}, nil
		}
	})
}

func TestDefaultService_CallReplicateAPI_ModelRegistry(t *testing.T) {
	ctx := context.Background()

	// Save original awsConfigLoader and restore after tests
	originalLoader := awsConfigLoader
	defer func() { awsConfigLoader = originalLoader }()

	// Mock AWS config loader
	awsConfigLoader = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{Region: "us-west-1"}, nil
	}

	t.Run("fail: invalid model in registry", func(t *testing.T) {
		// Create a service with the default model
		cfg := &ServiceConfig{
			BucketName:     "test-bucket",
			ReplicateToken: "test-token",
			ModelID:        model.ModelQwenImageEdit,
			S3Endpoint:     "http://localhost:9000",
			S3Region:       "us-west-1",
			S3AccessKey:    "test-key",
			S3SecretKey:    "test-secret",
			S3UsePathStyle: true,
			AppEnv:         "dev",
		}

		service, err := NewDefaultService(ctx, cfg)
		if err != nil {
			t.Fatalf("unexpected error creating service: %v", err)
		}

		// Manually set an invalid model ID to test error handling
		service.modelID = model.ModelID("invalid/model")

		// Try to call the API - should fail with model not found
		_, err = service.callReplicateAPI(ctx, "data:image/jpeg;base64,test", "test prompt", nil)
		if err == nil {
			t.Fatal("expected error for invalid model")
		}

		if err.Error() != "failed to get model metadata: model not found: invalid/model" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("fail: invalid input parameters", func(t *testing.T) {
		// Create a service with the default model
		cfg := &ServiceConfig{
			BucketName:     "test-bucket",
			ReplicateToken: "test-token",
			ModelID:        model.ModelQwenImageEdit,
			S3Endpoint:     "http://localhost:9000",
			S3Region:       "us-west-1",
			S3AccessKey:    "test-key",
			S3SecretKey:    "test-secret",
			S3UsePathStyle: true,
			AppEnv:         "dev",
		}

		service, err := NewDefaultService(ctx, cfg)
		if err != nil {
			t.Fatalf("unexpected error creating service: %v", err)
		}

		// Try to call the API with empty prompt - should fail validation
		_, err = service.callReplicateAPI(ctx, "data:image/jpeg;base64,test", "", nil)
		if err == nil {
			t.Fatal("expected error for empty prompt")
		}

		if err.Error() != "failed to build model input: prompt is required" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestDefaultService_BuildPrompt(t *testing.T) {
	ctx := context.Background()

	// Save original awsConfigLoader and restore after tests
	originalLoader := awsConfigLoader
	defer func() { awsConfigLoader = originalLoader }()

	// Mock AWS config loader
	awsConfigLoader = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{Region: "us-west-1"}, nil
	}

	cfg := &ServiceConfig{
		BucketName:     "test-bucket",
		ReplicateToken: "test-token",
		ModelID:        model.ModelQwenImageEdit,
		S3Endpoint:     "http://localhost:9000",
		S3Region:       "us-west-1",
		S3AccessKey:    "test-key",
		S3SecretKey:    "test-secret",
		S3UsePathStyle: true,
		AppEnv:         "dev",
	}

	service, err := NewDefaultService(ctx, cfg)
	if err != nil {
		t.Fatalf("unexpected error creating service: %v", err)
	}

	t.Run("success: builds prompt with default style", func(t *testing.T) {
		prompt := service.buildPrompt(nil, nil)

		if prompt == "" {
			t.Error("expected non-empty prompt")
		}

		// Check for key elements in the prompt
		if !contains(prompt, "modern") {
			t.Error("expected prompt to contain default style 'modern'")
		}
	})

	t.Run("success: builds prompt with custom style", func(t *testing.T) {
		style := "minimalist"
		prompt := service.buildPrompt(nil, &style)

		if !contains(prompt, "minimalist") {
			t.Error("expected prompt to contain custom style 'minimalist'")
		}
	})

	t.Run("success: builds prompt with room type", func(t *testing.T) {
		roomType := "living_room"
		prompt := service.buildPrompt(&roomType, nil)

		if !contains(prompt, "living_room") {
			t.Error("expected prompt to contain room type 'living_room'")
		}
	})

	t.Run("success: builds prompt with both room type and style", func(t *testing.T) {
		roomType := "bedroom"
		style := "rustic"
		prompt := service.buildPrompt(&roomType, &style)

		if !contains(prompt, "bedroom") {
			t.Error("expected prompt to contain room type 'bedroom'")
		}

		if !contains(prompt, "rustic") {
			t.Error("expected prompt to contain style 'rustic'")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
