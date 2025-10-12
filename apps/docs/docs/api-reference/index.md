# API Reference

Complete REST API documentation for Real Staging AI.

## OpenAPI Specification

The API is fully documented using OpenAPI 3.0 specification.

**Live API Documentation:**  
[https://jasonkradams.github.io/real-staging-ai/api/](https://jasonkradams.github.io/real-staging-ai/api/) - Interactive Swagger UI

**Local API Documentation:**  
[http://localhost:8080/api/v1/docs/](http://localhost:8080/api/v1/docs/) (when running locally)

## Base URL

```
Development: http://localhost:8080/api/v1
Production:  https://api.real-staging.ai/api/v1
```

## Authentication

All endpoints (except webhooks and health checks) require JWT authentication via Auth0.

**Header:**
```
Authorization: Bearer <your-jwt-token>
```

[Learn more about authentication →](../guides/authentication.md)

## Core Endpoints

### Projects

Manage user projects for organizing images.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/projects` | List all user projects |
| `POST` | `/projects` | Create a new project |
| `GET` | `/projects/{id}` | Get project details |
| `PATCH` | `/projects/{id}` | Update project |
| `DELETE` | `/projects/{id}` | Delete project |

### Uploads

Generate presigned URLs for direct S3 uploads.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/uploads/presign` | Get presigned upload URL |

### Images

Manage images and staging jobs.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/images` | Create image staging job |
| `POST` | `/images/batch` | Create multiple staging jobs |
| `GET` | `/images` | List images for a project |
| `GET` | `/images/{id}` | Get image details |
| `GET` | `/images/{id}/download` | Get presigned download URL |
| `DELETE` | `/images/{id}` | Delete image |

### Events (SSE)

Real-time updates via Server-Sent Events.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/events` | Subscribe to user events |

### Billing

Subscription and invoice management.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/billing/subscriptions` | Get user subscriptions |
| `GET` | `/billing/invoices` | List user invoices |

### Webhooks

Public endpoints for external integrations.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/stripe/webhook` | Stripe webhook handler |

### Health

Service health checks.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | API health status |

## Request Examples

### Create Project

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Downtown Condo Listings",
    "description": "Luxury condos for Q4 2024"
  }'
```

**Response (201 Created):**
```json
{
  "id": "01J9XYZ123ABC456DEF789GH",
  "name": "Downtown Condo Listings",
  "description": "Luxury condos for Q4 2024",
  "user_id": "user_abc123",
  "created_at": "2025-10-12T20:30:00Z",
  "updated_at": "2025-10-12T20:30:00Z"
}
```

### Request Presigned Upload URL

```bash
curl -X POST http://localhost:8080/api/v1/uploads/presign \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "01J9XYZ123ABC456DEF789GH",
    "filename": "living-room.jpg",
    "content_type": "image/jpeg"
  }'
```

**Response (200 OK):**
```json
{
  "upload_url": "https://s3.amazonaws.com/bucket/uploads/...",
  "key": "uploads/user_abc123/living-room-uuid.jpg",
  "expires_at": "2025-10-12T21:30:00Z"
}
```

### Create Image Staging Job

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "01J9XYZ123ABC456DEF789GH",
    "original_url": "s3://bucket/uploads/user_abc123/living-room-uuid.jpg",
    "room_type": "living_room",
    "style": "modern"
  }'
```

**Response (201 Created):**
```json
{
  "id": "01J9XYZ789ABC123DEF456GH",
  "project_id": "01J9XYZ123ABC456DEF789GH",
  "original_url": "s3://bucket/uploads/...",
  "staged_url": null,
  "room_type": "living_room",
  "style": "modern",
  "status": "queued",
  "created_at": "2025-10-12T20:32:00Z",
  "updated_at": "2025-10-12T20:32:00Z"
}
```

### Get Image Status

```bash
curl http://localhost:8080/api/v1/images/01J9XYZ789ABC123DEF456GH \
  -H "Authorization: Bearer $TOKEN"
```

**Response (200 OK - Ready):**
```json
{
  "id": "01J9XYZ789ABC123DEF456GH",
  "project_id": "01J9XYZ123ABC456DEF789GH",
  "original_url": "s3://bucket/uploads/...",
  "staged_url": "s3://bucket/staged/...",
  "room_type": "living_room",
  "style": "modern",
  "status": "ready",
  "processing_time_ms": 8945,
  "cost_cents": 1,
  "created_at": "2025-10-12T20:32:00Z",
  "updated_at": "2025-10-12T20:32:09Z"
}
```

## Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| `200` | OK | Request succeeded |
| `201` | Created | Resource created successfully |
| `204` | No Content | Request succeeded, no response body |
| `400` | Bad Request | Invalid request parameters |
| `401` | Unauthorized | Missing or invalid authentication |
| `403` | Forbidden | Insufficient permissions |
| `404` | Not Found | Resource not found |
| `422` | Unprocessable Entity | Validation error |
| `429` | Too Many Requests | Rate limit exceeded |
| `500` | Internal Server Error | Server error |
| `503` | Service Unavailable | Service temporarily unavailable |

## Error Responses

All errors return a consistent JSON structure:

```json
{
  "error": "validation_error",
  "message": "Invalid request parameters",
  "validation_errors": [
    {
      "field": "room_type",
      "message": "must be one of: living_room, bedroom, kitchen"
    }
  ]
}
```

## Rate Limiting

**Current Limits:**
- 100 requests per minute per user
- 1000 requests per hour per user

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1730000000
```

## Pagination

List endpoints support cursor-based pagination:

**Request:**
```bash
curl "http://localhost:8080/api/v1/images?limit=20&cursor=eyJpZCI6IjAxSjl...
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "next_cursor": "eyJpZCI6IjAxSjl...",
    "has_more": true
  }
}
```

## Webhooks

### Stripe Webhooks

Real Staging AI receives webhooks from Stripe for billing events.

**Endpoint:** `POST /api/v1/stripe/webhook`

**Events Handled:**
- `checkout.session.completed`
- `invoice.payment_succeeded`
- `invoice.payment_failed`
- `customer.subscription.updated`
- `customer.subscription.deleted`

[Learn more about webhooks →](../security/stripe-webhooks.md)

## SDKs & Tools

### Official SDKs

Coming soon:
- JavaScript/TypeScript SDK
- Python SDK
- Go SDK

### Postman Collection

Generate a Postman collection from OpenAPI spec:

```bash
make postman
# Creates postman_collection.json
```

Import into Postman and configure environment variables.

## Interactive Documentation

The API includes embedded Swagger UI for interactive exploration:

**Local:**
```
http://localhost:8080/api/v1/docs/
```

**Features:**
- Try API calls directly from browser
- See request/response examples
- Explore all endpoints and parameters
- Authentication testing

## Validation

All API inputs are validated with detailed error messages.

**Example Validation Rules:**

**Project Name:**
- Required
- 1-100 characters
- UTF-8 encoded

**File Upload:**
- Content-Type: `image/jpeg`, `image/png`, or `image/webp`
- Max size: 10 MB
- Valid image format

**Room Type:**
- One of: `living_room`, `bedroom`, `kitchen`, `bathroom`, `dining_room`, `office`

**Style:**
- One of: `modern`, `contemporary`, `traditional`, `scandinavian`, `industrial`, `bohemian`

## API Versioning

Current version: `v1`

Version is included in the URL path:
```
/api/v1/projects
```

Breaking changes will result in a new version (`v2`), with v1 maintained for backward compatibility.

---

**Full OpenAPI Specification:**  
[View on GitHub Pages →](https://jasonkradams.github.io/real-staging-ai/)

**Related Documentation:**
- [Authentication Guide](../guides/authentication.md)
- [Server-Sent Events](../guides/sse-events.md)
- [Your First Project Tutorial](../getting-started/first-project.md)
