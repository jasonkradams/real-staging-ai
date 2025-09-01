### Authentication
- **Auth0 (OIDC)**: SPA obtains access token; API validates JWT via JWKS
- Audience & issuer must match environment config

### Resources
- `POST /v1/uploads/presign` → Get S3 **presigned PUT** URL
- `POST /v1/images` → Create image + enqueue job (body: `{project_id, original_key, room_type, style}`)
- `GET /v1/images/{id}` → Fetch status/result
- `GET /v1/projects` / `POST /v1/projects` → basic project CRUD
- `POST /v1/stripe/webhook` → handle subscription events
- `GET /v1/events` → Server-Sent Events (image/job updates)

### OpenAPI Sketch
```yaml
openapi: 3.0.3
info:
  title: Virtual Staging API
  version: 0.1.0
servers:
  - url: https://api.local
paths:
  /v1/uploads/presign:
    post:
      security: [{ bearerAuth: [] }]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                contentType: { type: string }
                projectId: { type: string }
      responses:
        '200':
          description: presigned upload
          content:
            application/json:
              schema:
                type: object
                properties:
                  url: { type: string }
                  key: { type: string }
  /v1/images:
    post:
      security: [{ bearerAuth: [] }]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [projectId, originalKey]
              properties:
                projectId: { type: string }
                originalKey: { type: string }
                roomType: { type: string }
                style: { type: string }
      responses:
        '202':
          description: accepted
          content:
            application/json:
              schema:
                type: object
                properties:
                  id: { type: string }
  /v1/images/{id}:
    get:
      security: [{ bearerAuth: [] }]
      parameters:
        - in: path
          name: id
          required: true
          schema: { type: string }
      responses:
        '200':
          description: image
          content:
            application/json:
              schema:
                type: object
                properties:
                  id: { type: string }
                  status: { type: string, enum: [queued, processing, ready, error] }
                  originalUrl: { type: string }
                  stagedUrl: { type: string, nullable: true }
                  error: { type: string, nullable: true }
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```
