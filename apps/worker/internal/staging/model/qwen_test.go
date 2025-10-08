package model

import (
	"context"
	"testing"
)

func TestNewQwenInputBuilder(t *testing.T) {
	t.Run("success: creates new builder", func(t *testing.T) {
		builder := NewQwenInputBuilder()

		if builder == nil {
			t.Fatal("expected builder to be non-nil")
		}
	})
}

func TestQwenInputBuilder_BuildInput(t *testing.T) {
	ctx := context.Background()

	t.Run("success: builds input without seed", func(t *testing.T) {
		builder := NewQwenInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			Prompt:       "Add modern furniture",
		}

		input, err := builder.BuildInput(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify required fields
		if input["image"] != req.ImageDataURL {
			t.Errorf("expected image to be %s, got %v", req.ImageDataURL, input["image"])
		}

		if input["prompt"] != req.Prompt {
			t.Errorf("expected prompt to be %s, got %v", req.Prompt, input["prompt"])
		}

		// Verify default parameters
		if input["go_fast"] != true {
			t.Errorf("expected go_fast to be true, got %v", input["go_fast"])
		}

		if input["aspect_ratio"] != "match_input_image" {
			t.Errorf("expected aspect_ratio to be 'match_input_image', got %v", input["aspect_ratio"])
		}

		if input["output_format"] != "webp" {
			t.Errorf("expected output_format to be 'webp', got %v", input["output_format"])
		}

		if input["output_quality"] != 80 {
			t.Errorf("expected output_quality to be 80, got %v", input["output_quality"])
		}

		// Verify seed is not set
		if _, exists := input["seed"]; exists {
			t.Error("expected seed to not be set")
		}
	})

	t.Run("success: builds input with seed", func(t *testing.T) {
		builder := NewQwenInputBuilder()
		seed := int64(12345)
		req := &ModelInputRequest{
			ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			Prompt:       "Add modern furniture",
			Seed:         &seed,
		}

		input, err := builder.BuildInput(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify seed is set
		if input["seed"] != seed {
			t.Errorf("expected seed to be %d, got %v", seed, input["seed"])
		}
	})

	t.Run("fail: nil request", func(t *testing.T) {
		builder := NewQwenInputBuilder()

		_, err := builder.BuildInput(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil request")
		}

		expectedMsg := "request cannot be nil"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("fail: empty image data URL", func(t *testing.T) {
		builder := NewQwenInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "",
			Prompt:       "Add modern furniture",
		}

		_, err := builder.BuildInput(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty image data URL")
		}

		expectedMsg := "image data URL is required"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("fail: empty prompt", func(t *testing.T) {
		builder := NewQwenInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			Prompt:       "",
		}

		_, err := builder.BuildInput(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty prompt")
		}

		expectedMsg := "prompt is required"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})
}

func TestQwenInputBuilder_Validate(t *testing.T) {
	t.Run("success: valid request", func(t *testing.T) {
		builder := NewQwenInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			Prompt:       "Add modern furniture",
		}

		err := builder.Validate(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success: valid request with seed", func(t *testing.T) {
		builder := NewQwenInputBuilder()
		seed := int64(12345)
		req := &ModelInputRequest{
			ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			Prompt:       "Add modern furniture",
			Seed:         &seed,
		}

		err := builder.Validate(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("fail: nil request", func(t *testing.T) {
		builder := NewQwenInputBuilder()

		err := builder.Validate(nil)
		if err == nil {
			t.Fatal("expected error for nil request")
		}
	})

	t.Run("fail: empty image data URL", func(t *testing.T) {
		builder := NewQwenInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "",
			Prompt:       "Add modern furniture",
		}

		err := builder.Validate(req)
		if err == nil {
			t.Fatal("expected error for empty image data URL")
		}
	})

	t.Run("fail: empty prompt", func(t *testing.T) {
		builder := NewQwenInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			Prompt:       "",
		}

		err := builder.Validate(req)
		if err == nil {
			t.Fatal("expected error for empty prompt")
		}
	})
}
