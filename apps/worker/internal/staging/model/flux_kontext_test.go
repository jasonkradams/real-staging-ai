package model

import (
	"context"
	"testing"
)

func TestNewFluxKontextInputBuilder(t *testing.T) {
	t.Run("success: creates new builder", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()

		if builder == nil {
			t.Fatal("expected builder to be non-nil")
		}
	})
}

func TestFluxKontextInputBuilder_BuildInput(t *testing.T) {
	ctx := context.Background()

	t.Run("success: builds input without seed and without image", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()
		req := &ModelInputRequest{
			Prompt: "Create a modern living room",
		}

		input, err := builder.BuildInput(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify required fields
		if input["prompt"] != req.Prompt {
			t.Errorf("expected prompt to be %s, got %v", req.Prompt, input["prompt"])
		}

		// Verify default parameters
		if input["aspect_ratio"] != "match_input_image" {
			t.Errorf("expected aspect_ratio to be 'match_input_image', got %v", input["aspect_ratio"])
		}

		if input["output_format"] != "png" {
			t.Errorf("expected output_format to be 'png', got %v", input["output_format"])
		}

		if input["safety_tolerance"] != 2 {
			t.Errorf("expected safety_tolerance to be 2, got %v", input["safety_tolerance"])
		}

		if input["prompt_upsampling"] != false {
			t.Errorf("expected prompt_upsampling to be false, got %v", input["prompt_upsampling"])
		}

		// Verify seed is not set
		if _, exists := input["seed"]; exists {
			t.Error("expected seed to not be set")
		}
	})

	t.Run("success: builds input with image", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			Prompt:       "Add modern furniture",
		}

		input, err := builder.BuildInput(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify input_image is set
		if input["input_image"] != req.ImageDataURL {
			t.Errorf("expected input_image to be %s, got %v", req.ImageDataURL, input["input_image"])
		}
	})

	t.Run("success: builds input with seed", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()
		seed := int64(12345)
		req := &ModelInputRequest{
			Prompt: "Add modern furniture",
			Seed:   &seed,
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

	t.Run("success: builds input with all parameters", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()
		seed := int64(99999)
		req := &ModelInputRequest{
			ImageDataURL: "data:image/png;base64,iVBORw0KGgo=",
			Prompt:       "Transform this room into a cozy bedroom",
			Seed:         &seed,
		}

		input, err := builder.BuildInput(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify all fields are set correctly
		if input["prompt"] != req.Prompt {
			t.Errorf("expected prompt to be %s, got %v", req.Prompt, input["prompt"])
		}
		if input["input_image"] != req.ImageDataURL {
			t.Errorf("expected input_image to be %s, got %v", req.ImageDataURL, input["input_image"])
		}
		if input["seed"] != seed {
			t.Errorf("expected seed to be %d, got %v", seed, input["seed"])
		}
	})

	t.Run("fail: nil request", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()

		_, err := builder.BuildInput(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil request")
		}

		expectedMsg := "request cannot be nil"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("fail: empty prompt", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()
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

func TestFluxKontextInputBuilder_Validate(t *testing.T) {
	t.Run("success: valid request with prompt only", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()
		req := &ModelInputRequest{
			Prompt: "Create a modern kitchen",
		}

		err := builder.Validate(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success: valid request with image and prompt", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			Prompt:       "Add modern furniture",
		}

		err := builder.Validate(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success: valid request with all fields", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()
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
		builder := NewFluxKontextInputBuilder()

		err := builder.Validate(nil)
		if err == nil {
			t.Fatal("expected error for nil request")
		}

		expectedMsg := "request cannot be nil"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("fail: empty prompt", func(t *testing.T) {
		builder := NewFluxKontextInputBuilder()
		req := &ModelInputRequest{
			ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			Prompt:       "",
		}

		err := builder.Validate(req)
		if err == nil {
			t.Fatal("expected error for empty prompt")
		}

		expectedMsg := "prompt is required"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})
}
