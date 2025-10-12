# OpenAPI Specification

The complete OpenAPI 3.0 specification for the Real Staging AI API.

## Live Documentation

**Hosted Documentation:**  
[https://jasonkradams.github.io/real-staging-ai/](https://jasonkradams.github.io/real-staging-ai/)

**Local Development:**  
[http://localhost:8080/api/v1/docs/](http://localhost:8080/api/v1/docs/)

## Specification File

The OpenAPI specification is maintained in:
```
apps/api/web/api/v1/oas3.yaml
```

## Validation

Validate the OpenAPI spec:

```bash
make docs
```

This runs:
```bash
docker run --rm -v $(CURDIR)/apps/api/web/api/v1:/spec \
  python:3.13-slim /bin/sh -c \
  "pip install openapi-spec-validator && openapi-spec-validator /spec/oas3.yaml"
```

## Generating Postman Collection

Convert OpenAPI spec to Postman collection:

```bash
make postman
```

Output: `postman_collection.json`

## Key Features

### Authentication

All endpoints (except public webhooks) require JWT authentication:

```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []
```

### Request Validation

Strict validation on all inputs:

```yaml
components:
  schemas:
    CreateProjectRequest:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 100
        description:
          type: string
          maxLength: 500
```

### Response Examples

Complete examples for all endpoints:

```yaml
responses:
  '201':
    description: Project created successfully
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Project'
        example:
          id: "01J9XYZ123ABC456DEF789GH"
          name: "My Project"
          created_at: "2025-10-12T20:30:00Z"
```

### Error Responses

Consistent error structure:

```yaml
components:
  schemas:
    ErrorResponse:
      type: object
      properties:
        error:
          type: string
        message:
          type: string
        validation_errors:
          type: array
          items:
            $ref: '#/components/schemas/ValidationError'
```

## Spec Structure

```yaml
openapi: 3.0.3
info:
  title: Real Staging AI API
  version: 1.0.0
  description: Professional AI-powered virtual staging

servers:
  - url: http://localhost:8080/api/v1
    description: Development
  - url: https://api.real-staging.ai/api/v1
    description: Production

paths:
  /projects:
    get: ...
    post: ...
  /projects/{id}:
    get: ...
    patch: ...
    delete: ...
  # ... more endpoints

components:
  schemas:
    Project: ...
    Image: ...
    # ... more schemas
  
  securitySchemes:
    bearerAuth: ...

security:
  - bearerAuth: []
```

## Code Generation

The spec can be used for:

- **Server stubs** - Generate server code (Go, Python, Node.js)
- **Client SDKs** - Generate client libraries
- **Documentation** - Generate HTML/Markdown docs
- **Mock servers** - Create mock API for testing

### Example: Generate Go Server

```bash
openapi-generator-cli generate \
  -i apps/api/web/api/v1/oas3.yaml \
  -g go-server \
  -o generated/server
```

### Example: Generate TypeScript Client

```bash
openapi-generator-cli generate \
  -i apps/api/web/api/v1/oas3.yaml \
  -g typescript-fetch \
  -o generated/client
```

## Contributing

When adding new endpoints:

1. Update `oas3.yaml` with new paths and schemas
2. Validate spec: `make docs`
3. Test in Swagger UI locally
4. Include request/response examples
5. Document all parameters and fields
6. Add appropriate error responses

## Resources

- [OpenAPI Specification](https://spec.openapis.org/oas/v3.0.3)
- [Swagger Editor](https://editor.swagger.io/)
- [OpenAPI Generator](https://openapi-generator.tech/)

---

[← Back to API Reference](index.md)
