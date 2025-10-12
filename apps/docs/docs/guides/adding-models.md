# Adding a New AI Model

This guide walks through the process of adding a new AI model to the Real Staging system.

## Prerequisites

- Understanding of the [Model Registry Architecture](../model_registry.md)
- Access to the new model's API documentation
- Go development environment set up

## Step-by-Step Process

### 1. Define the Model ID Constant

Add a new constant to `apps/worker/internal/staging/model/registry.go`:

```go
const (
    ModelQwenImageEdit  ModelID = "qwen/qwen-image-edit"
    ModelFluxKontextMax ModelID = "black-forest-labs/flux-kontext-max"
    ModelYourNewModel   ModelID = "vendor/model-name"  // Add your model here
)
```

### 2. Create the Input Builder

Create a new file `apps/worker/internal/staging/model/yourmodel.go`:

```go
package staging

import (
    "context"
    "fmt"
    
    "github.com/replicate/replicate-go"
)

// YourModelInputBuilder builds input parameters for Your Model.
type YourModelInputBuilder struct{}

// Ensure YourModelInputBuilder implements ModelInputBuilder.
var _ ModelInputBuilder = (*YourModelInputBuilder)(nil)

// NewYourModelInputBuilder creates a new YourModelInputBuilder.
func NewYourModelInputBuilder() *YourModelInputBuilder {
    return &YourModelInputBuilder{}
}

// BuildInput creates the input parameters for Your Model.
func (b *YourModelInputBuilder) BuildInput(ctx context.Context, req *ModelInputRequest) (replicate.PredictionInput, error) {
    if err := b.Validate(req); err != nil {
        return nil, err
    }

    // Build input according to your model's API contract
    input := replicate.PredictionInput{
        "image":  req.ImageDataURL,
        "prompt": req.Prompt,
        // Add model-specific parameters here
    }

    if req.Seed != nil {
        input["seed"] = *req.Seed
    }

    return input, nil
}

// Validate checks if the request is valid for Your Model.
func (b *YourModelInputBuilder) Validate(req *ModelInputRequest) error {
    if req == nil {
        return fmt.Errorf("request cannot be nil")
    }
    if req.ImageDataURL == "" {
        return fmt.Errorf("image data URL is required")
    }
    if req.Prompt == "" {
        return fmt.Errorf("prompt is required")
    }
    // Add model-specific validation here
    return nil
}
```

### 3. Register the Model

Update `NewModelRegistry()` in `apps/worker/internal/staging/model/registry.go`:

```go
func NewModelRegistry() *ModelRegistry {
    registry := &ModelRegistry{
        models: make(map[ModelID]*ModelMetadata),
    }

    // Existing models...
    registry.Register(&ModelMetadata{
        ID:           ModelQwenImageEdit,
        Name:         "Qwen Image Edit",
        Description:  "Fast image editing model optimized for virtual staging",
        Version:      "latest",
        InputBuilder: NewQwenInputBuilder(),
    })

    // Add your new model
    registry.Register(&ModelMetadata{
        ID:           ModelYourNewModel,
        Name:         "Your Model Name",
        Description:  "Description of what your model does",
        Version:      "v1.0",
        InputBuilder: NewYourModelInputBuilder(),
    })

    return registry
}
```

### 4. Write Comprehensive Tests

Create `apps/worker/internal/staging/model/yourmodel_test.go`:

```go
package staging

import (
    "context"
    "testing"
)

func TestNewYourModelInputBuilder(t *testing.T) {
    t.Run("success: creates new builder", func(t *testing.T) {
        builder := NewYourModelInputBuilder()
        if builder == nil {
            t.Fatal("expected builder to be non-nil")
        }
    })
}

func TestYourModelInputBuilder_BuildInput(t *testing.T) {
    ctx := context.Background()

    t.Run("success: builds input without seed", func(t *testing.T) {
        builder := NewYourModelInputBuilder()
        req := &ModelInputRequest{
            ImageDataURL: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
            Prompt:       "Add modern furniture",
        }

        input, err := builder.BuildInput(ctx, req)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        // Verify all required fields
        if input["image"] != req.ImageDataURL {
            t.Error("image field mismatch")
        }
        // Add more assertions...
    })

    // Add tests for:
    // - success: builds input with seed
    // - fail: nil request
    // - fail: empty image data URL
    // - fail: empty prompt
    // - fail: any model-specific validation errors
}

func TestYourModelInputBuilder_Validate(t *testing.T) {
    // Add comprehensive validation tests
    // Must cover all success and failure cases
}
```

### 5. Run Tests

Ensure 100% test coverage for your new code:

```bash
cd apps/worker/internal/staging/model
go test -v -cover .
```

Target coverage should be 100% for all new code (model and builder files). The model package currently achieves 100% test coverage.

### 6. Update Documentation

Add your model to `docs/model_registry.md`:

```markdown
### 2. Your Model Name

- **ID**: `vendor/model-name`
- **Description**: Description of what your model does
- **Parameters**:
  - `image` (string, required): Base64-encoded image data URL
  - `prompt` (string, required): Editing instructions
  - `your_param` (type): Description
  - `seed` (int, optional): Random seed for reproducibility
```

### 7. Optional: Update Model Selection

To make your model the default, update `apps/worker/main.go`:

```go
import (
    "github.com/real-staging-ai/worker/internal/staging"
    "github.com/real-staging-ai/worker/internal/staging/model"
)

stagingCfg := &staging.ServiceConfig{
    // ...
    ModelID: model.ModelYourNewModel, // Change default here
    // ...
}
```

Or allow runtime selection via configuration or admin UI (future enhancement).

## Testing Your Model

### Unit Tests

Run the unit tests to verify your implementation:

```bash
cd apps/worker/internal/staging
go test -v -run TestYourModel
```

### Integration Testing

To test with the actual Replicate API:

1. Set up environment variables:
   ```bash
   export REPLICATE_API_TOKEN=your_token
   export S3_BUCKET_NAME=your_bucket
   ```

2. Create a test staging request
3. Verify the output meets expectations

### Manual Testing

1. Update worker config to use your model
2. Start the worker: `make worker`
3. Enqueue a staging job via the API
4. Monitor logs and verify results

## Common Issues

### Model Not Found Error

If you see `model not found: vendor/model-name`:
- Verify the model ID constant matches the registration
- Ensure `NewModelRegistry()` is called before using the model
- Check for typos in the model ID

### Input Validation Errors

If inputs are rejected:
- Review the model's API documentation
- Verify all required fields are included
- Check data types match the API contract

### API Errors from Replicate

If the Replicate API returns errors:
- Verify the model ID exists on Replicate
- Check parameter names match the API exactly
- Ensure your Replicate token has access to the model

## Best Practices

1. **Keep It Simple**: Start with minimal required parameters
2. **Validate Early**: Add validation in the builder, not in the service
3. **Document Well**: Include Godoc comments on all public types/functions
4. **Test Thoroughly**: Aim for 100% coverage of new code
5. **Follow Conventions**: Use existing builders as templates
6. **Error Messages**: Make validation errors clear and actionable

## Future Enhancements

- Admin UI for model selection and configuration
- Per-project model preferences
- A/B testing support
- Cost tracking per model
- Performance metrics and comparison

## Questions?

- Review existing models: 
  - `apps/worker/internal/staging/model/qwen.go`
  - `apps/worker/internal/staging/model/flux_kontext.go`
- Check the architecture doc: `docs/model_registry.md`
- Ask in the team channel or create an issue
