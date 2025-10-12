### Authentication
- **Auth0 (OIDC)**: SPA obtains access token; API validates JWT via JWKS
- Audience & issuer must match environment config

### Resources
- `POST /api/v1/uploads/presign` → Get S3 **presigned PUT** URL (body: `{filename, content_type, file_size}`)
- `POST /api/v1/images` → Create image + enqueue job (body: `{project_id, original_url, room_type?, style?, seed?}`)
- `GET /api/v1/images/{id}` → Fetch status/result
- `GET /api/v1/projects` / `POST /api/v1/projects` → basic project CRUD
- `POST /api/v1/stripe/webhook` → handle subscription events
- `GET /api/v1/events?image_id={id}` → Server-Sent Events (per-image channel; minimal payload with status only)

### OpenAPI Sketch
```yaml
openapi: 3.0.3
info:
  title: Real Staging API
  version: 0.1.0
servers:
  - url: https://api.local
paths:
  /api/v1/uploads/presign:
    post:
      security: [{ bearerAuth: [] }]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                filename: { type: string }
                content_type: { type: string }
                file_size: { type: integer }
      responses:
        '200':
          description: presigned upload
          content:
            application/json:
              schema:
                type: object
                properties:
                  upload_url: { type: string }
                  file_key: { type: string }
                  expires_in: { type: integer }
  /api/v1/images:
    post:
      security: [{ bearerAuth: [] }]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [project_id, original_url]
              properties:
                project_id: { type: string }
                original_url: { type: string }
                room_type: { type: string }
                style: { type: string }
                seed: { type: integer }
      responses:
        '202':
          description: accepted
          content:
            application/json:
              schema:
                type: object
                properties:
                  id: { type: string }
  /api/v1/images/{id}:
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
  /api/v1/events:
    get:
      security: [{ bearerAuth: [] }]
      parameters:
        - in: query
          name: image_id
          required: true
          schema: { type: string }
      responses:
        '200':
          description: server-sent events stream (per-image; status-only payload)
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```
