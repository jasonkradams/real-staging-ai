# Staging Model Registry Migration

**Date**: 2025-10-07  
**Status**: Completed

## Overview

The staging service has been refactored to support multiple AI models through a registry-based architecture. Previously, the model was hardcoded with configuration in YAML files. Now, models are defined as code constants with their own API contract implementations.

## Changes Made

### New Files

1. **`apps/worker/internal/staging/model.go`**
   - Defines `ModelID` enum type
   - Implements `ModelRegistry` for managing available models
   - Defines `ModelInputBuilder` interface for model-specific inputs
   - Exports `ModelQwenImageEdit` constant

2. **`apps/worker/internal/staging/qwen_input_builder.go`**
   - Implements `QwenInputBuilder` for Qwen Image Edit model
   - Handles model-specific input parameters and validation

3. **`apps/worker/internal/staging/model_test.go`**
   - Comprehensive tests for model registry (100% coverage)
   - Tests for registration, retrieval, listing, and existence checks

4. **`apps/worker/internal/staging/qwen_input_builder_test.go`**
   - Full test coverage for Qwen input builder
   - Tests all success and failure scenarios

5. **`apps/worker/internal/staging/default_service_test.go`**
   - Tests for `DefaultService` model registry integration
   - Tests for config validation and error handling

6. **`docs/model_registry.md`**
   - Architecture documentation for the model registry system
   - Explains design decisions and usage patterns

7. **`docs/guides/ADDING_NEW_MODEL.md`**
   - Step-by-step guide for adding new models
   - Includes code examples and best practices

### Modified Files

1. **`apps/worker/internal/staging/default_service.go`**
   - Changed `modelVersion string` to `modelID ModelID`
   - Added `registry *ModelRegistry` field
   - Updated `ServiceConfig` to use `ModelID` instead of `ModelVersion`
   - Modified `NewDefaultService()` to initialize and validate model registry
   - Refactored `callReplicateAPI()` to use model input builders
   - Removed hardcoded input parameters for Qwen model

2. **`apps/worker/main.go`**
   - Updated staging config initialization to use `ModelID`
   - Changed from `ModelVersion: cfg.Replicate.ModelVersion` to `ModelID: staging.ModelQwenImageEdit`

3. **`apps/worker/internal/config/config.go`**
   - Removed `ModelVersion` field from `Replicate` struct
   - Added comment explaining the change

4. **`config/shared.yml`**
   - Removed `model_version: qwen/qwen-image-edit` configuration
   - Added comment explaining model selection is now in code

5. **`docs/worker_service.md`**
   - Added "AI Model Registry" section
   - Documented currently supported models
   - Referenced architecture documentation

## Migration Guide

### For Developers

No immediate action required. The changes are backward compatible:

- Default model remains `qwen/qwen-image-edit`
- No database schema changes
- No API contract changes
- No changes to job payloads

### For Deployments

#### Configuration Changes

**Before** (config/shared.yml):
```yaml
replicate:
  model_version: qwen/qwen-image-edit
```

**After** (config/shared.yml):
```yaml
replicate:
  # Model selection is now handled in code via staging.ModelID enum
```

#### Environment Variables

The `REPLICATE_MODEL_VERSION` environment variable is no longer used. If set, it will be ignored.

#### No Downtime Required

These changes can be deployed without downtime:
1. The worker will continue using Qwen Image Edit by default
2. No database migrations needed
3. No API version changes

### For Future Development

When adding a new model:

1. Follow the guide: `docs/guides/ADDING_NEW_MODEL.md`
2. Define a new `ModelID` constant
3. Implement `ModelInputBuilder` interface
4. Register in `NewModelRegistry()`
5. Write comprehensive tests
6. Update documentation

## Testing

All new code has comprehensive test coverage:

```bash
cd apps/worker/internal/staging
go test -v -cover .
```

**Coverage**: 33.0% of statements (new model registry code is fully covered)

Tests include:
- ✅ Model registry operations
- ✅ Input builder validation
- ✅ Service initialization with models
- ✅ Error handling for invalid models
- ✅ Default model behavior

## Benefits

1. **Type Safety**: Models are now compile-time constants, not runtime strings
2. **Extensibility**: Easy to add new models without changing core logic
3. **Testability**: Each model's input logic is independently testable
4. **Maintainability**: Clear separation between models and service logic
5. **Documentation**: Self-documenting through registry metadata

## Future Enhancements

### Admin UI for Model Selection

A future enhancement will allow admins to:
- Select active model from registry
- Configure model-specific parameters
- Set per-project model preferences
- View model metadata and capabilities

This will be implemented as:
1. New API endpoint: `GET /admin/models` (list available models)
2. New API endpoint: `PUT /admin/settings/model` (update active model)
3. Admin UI page for model configuration
4. Role-based access control (admin only)

### Per-Project Models

Allow different projects to use different models:
- Store `model_id` in projects table
- Pass model ID in job payload
- Initialize service with project's preferred model

### A/B Testing

Support running multiple models for comparison:
- Split traffic between models
- Track quality metrics per model
- Automatic quality scoring

### Cost Tracking

Track API costs by model:
- Log model used for each job
- Aggregate costs in analytics
- Budget alerts per model

## Rollback Plan

If issues arise, rollback is simple:

1. Revert to previous commit
2. Restore `model_version` in config files
3. Redeploy worker

No database changes were made, so rollback is safe.

## Related Documents

- [Model Registry Architecture](./model_registry.md)
- [Adding a New Model Guide](./guides/ADDING_NEW_MODEL.md)
- [Worker Service Documentation](./worker_service.md)
- [Repository Guidelines](../AGENTS.md)

## Questions or Issues?

- Check the architecture doc: `docs/model_registry.md`
- Review the test files for usage examples
- Create an issue in the repository
