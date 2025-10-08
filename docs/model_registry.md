# Model Registry Architecture

This document describes the model registry system used by the staging service to support multiple AI models with different API contracts.

## Overview

The staging service supports multiple Replicate AI models for virtual staging. Each model has its own API contract (input parameters), which is handled through a model registry system.

## Architecture

### Package Structure

The model registry is now organized in a dedicated package:
- **Location**: `apps/worker/internal/staging/model/`
- **Files**:
  - `registry.go` - Core registry and interface definitions
  - `qwen.go` - Qwen Image Edit model implementation
  - `flux_kontext.go` - Flux Kontext Max model implementation
  - `*_test.go` - Comprehensive test files (100% coverage)

### Components

1. **Model Enum**: A predefined enum of supported models
2. **ModelInputBuilder Interface**: Defines how to build input parameters for each model
3. **Model Registry**: Maps model IDs to their input builders and metadata
4. **DefaultService**: Uses the registry to build appropriate inputs for the selected model

### Model Enum

Models are defined as constants in the codebase (not configuration files). This ensures type safety and prevents runtime errors from invalid model names.

```go
type ModelID string

const (
    ModelQwenImageEdit  ModelID = "qwen/qwen-image-edit"
    ModelFluxKontextMax ModelID = "black-forest-labs/flux-kontext-max"
    // Additional models can be added here
)
```

### ModelInputBuilder Interface

Each model has a specific API contract. The `ModelInputBuilder` interface abstracts the creation of model-specific input parameters:

```go
type ModelInputBuilder interface {
    // BuildInput creates the input parameters for the model
    BuildInput(ctx context.Context, req *ModelInputRequest) (replicate.PredictionInput, error)
    
    // Validate checks if the request is valid for this model
    Validate(req *ModelInputRequest) error
}
```

### Model Metadata

Each registered model includes metadata:

- **ID**: Unique identifier (e.g., "qwen/qwen-image-edit")
- **Name**: Human-readable name
- **Description**: Brief description of the model's capabilities
- **Version**: Model version for tracking
- **InputBuilder**: Implementation of ModelInputBuilder for this model

## Supported Models

### 1. Qwen Image Edit

- **ID**: `qwen/qwen-image-edit`
- **Description**: Fast image editing model optimized for staging
- **Package Location**: `apps/worker/internal/staging/model/qwen.go`
- **Parameters**:
  - `image` (string, required): Base64-encoded image data URL
  - `prompt` (string, required): Editing instructions
  - `go_fast` (bool): Enable fast mode (default: true)
  - `aspect_ratio` (string): Output aspect ratio (default: "match_input_image")
  - `output_format` (string): Output format (default: "webp")
  - `output_quality` (int): Output quality 1-100 (default: 80)
  - `seed` (int, optional): Random seed for reproducibility

### 2. Flux Kontext Max

- **ID**: `black-forest-labs/flux-kontext-max`
- **Description**: High-quality image generation and editing with advanced context understanding
- **Package Location**: `apps/worker/internal/staging/model/flux_kontext.go`
- **Parameters**:
  - `prompt` (string, required): Text description or editing instruction
  - `input_image` (string, optional): Base64-encoded image data URL for image editing
  - `aspect_ratio` (string): Output aspect ratio (default: "match_input_image")
  - `output_format` (string): Output format - "jpg" or "png" (default: "png")
  - `safety_tolerance` (int): Safety level 0-6, 2 is max with input images (default: 2)
  - `prompt_upsampling` (bool): Automatic prompt improvement (default: false)
  - `seed` (int, optional): Random seed for reproducibility

### Future Models

Additional models can be added by:

1. Defining a new constant in the ModelID enum
2. Creating a new implementation of ModelInputBuilder
3. Registering the model in the registry with its metadata
4. Adding documentation to this file

## Configuration

### Current Approach (Being Replaced)

Previously, the model version was stored in `config/shared.yml`:

```yaml
replicate:
  model_version: qwen/qwen-image-edit
```

### New Approach

Models are now defined in code as constants. Model selection will eventually be exposed through an admin-only configuration UI, allowing:

- Selection of the active model
- Configuration of model-specific parameters
- Per-project or global model settings

## Usage

### Initializing the Service

```go
import (
    "github.com/virtual-staging-ai/worker/internal/staging"
    "github.com/virtual-staging-ai/worker/internal/staging/model"
)

// Create service with specific model
stagingService, err := staging.NewDefaultService(ctx, &staging.ServiceConfig{
    ModelID:        model.ModelQwenImageEdit,     // or model.ModelFluxKontextMax
    BucketName:     cfg.S3Bucket(),
    ReplicateToken: cfg.Replicate.APIToken,
    // ... other config
})
```

### Staging an Image

The service automatically uses the correct input builder for the configured model:

```go
stagedURL, err := stagingService.StageImage(ctx, &staging.StagingRequest{
    ImageID:     "img-123",
    OriginalURL: "s3://bucket/uploads/original.jpg",
    RoomType:    ptr("living_room"),
    Style:       ptr("modern"),
    Seed:        ptr(int64(12345)),
})
```

## Testing

Each model input builder must have:

- Unit tests with 100% coverage
- Tests for all input parameters
- Validation error tests
- Integration tests with mock Replicate API

## Future Enhancements

1. **Admin UI**: Web interface for model selection and configuration
2. **Per-Project Models**: Allow different models per project
3. **A/B Testing**: Support running multiple models for comparison
4. **Cost Tracking**: Track API costs per model
5. **Performance Metrics**: Monitor quality and speed by model
6. **Model Fallback**: Automatic fallback if primary model fails

## Migration Notes

When migrating from the old system:

1. Remove `model_version` from config files
2. Update worker initialization to use `ModelID` constant
3. The default model remains `ModelQwenImageEdit`
4. No changes required to the database or API contracts
