# Implementation Notes

Technical documentation about key features and architectural decisions implemented in Real Staging AI.

## Overview

This section contains detailed implementation notes, migration guides, and technical deep-dives for specific features. These documents provide context about why certain decisions were made and how complex features work under the hood.

## Available Documentation

### [Multi-Upload Implementation](multi-upload.md)

Details the batch upload feature allowing users to upload up to 50 images at once with customizable settings per image.

**Key Topics:**
- Batch presigned URL generation
- Frontend drag-and-drop interface
- Default vs per-image configuration
- Job queue optimization

### [Configuration Migration](configuration-migration.md)

Documents the migration from environment variables to structured YAML configuration files with secrets separation.

**Key Topics:**
- YAML-based configuration structure
- Secrets management
- Environment-specific configs
- Migration guide from old format

### [Model Package Refactor](model-refactors.md)

Explains the refactoring of AI model integration to support multiple models through a registry pattern.

**Key Topics:**
- Model registry architecture
- Input builder pattern
- Model metadata structure
- Adding new models

### [Staging Model Registry Migration](staging-model-registry.md)

Details the migration to a centralized model registry system for managing multiple AI staging models.

**Key Topics:**
- Registry design
- Model selection logic
- Input validation per model
- Testing strategies

## When to Read These

**You should read these documents if you:**

- Want to understand how a specific feature works in detail
- Need to extend or modify existing features
- Are debugging issues related to these features
- Want context on architectural decisions
- Are implementing similar features

**You can skip these if you:**

- Just want to use the API (see [API Reference](../api-reference/))
- Are getting started (see [Getting Started](../getting-started/))
- Need operational guides (see [Operations](../operations/))

## Document Structure

Each implementation note follows this structure:

1. **Background** - Why was this needed?
2. **Goals** - What were we trying to achieve?
3. **Implementation** - How was it built?
4. **Usage** - How to use the feature
5. **Testing** - How it's tested
6. **Future Work** - Potential improvements

## Related Documentation

- [Architecture](../architecture/) - System design overview
- [Development](../development/) - Contributing guidelines
- [Guides](../guides/) - How-to guides for common tasks

---

**Questions?** Open an issue or start a discussion on GitHub.
