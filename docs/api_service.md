# API Service

This document provides a detailed documentation of the API service.

## Authentication

The API service uses Auth0 for authentication. All protected endpoints require a valid JWT token in the `Authorization` header as a Bearer token.

The authentication flow is as follows:

1.  The user authenticates with Auth0 and obtains a JWT token.
2.  The user sends a request to a protected endpoint with the JWT token in the `Authorization` header.
3.  The API service validates the token using the public key from the Auth0 JWKS endpoint.
4.  If the token is valid, the API service extracts the user's `sub` claim (Auth0 user ID) and uses it to identify the user.
5.  If a user with the given `auth0_sub` does not exist in the database, a new user is created.

## HTTP Handlers

The API service exposes the following groups of HTTP handlers:

- **Projects (`/api/v1/projects`):** Handles CRUD operations for projects.
- **Uploads (`/api/v1/uploads`):** Handles file uploads, including generating presigned URLs for S3.
- **Images (`/api/v1/images`):** Handles CRUD operations for images.
- **Events (`/api/v1/events`):** Handles Server-Sent Events (SSE) for real-time updates.
- **Stripe (`/api/v1/stripe/webhook`):** Handles webhooks from Stripe for payment processing (public route, signature-verified, idempotent).

## Database Interaction

The API service uses the `pgx` library to interact with the PostgreSQL database. It uses a connection pool to manage database connections.

The database access layer is organized into repositories, which encapsulate the SQL queries for each database table.
The SQL queries are defined in `.sql` files and `sqlc` is used to generate type-safe Go code from them.

## S3 Usage

The API service uses an S3-compatible object storage service (MinIO in the development environment) for storing user-uploaded images and other large files.

When a user wants to upload a file, the API service generates a presigned URL that allows the client to upload the file directly to the S3 bucket. This avoids proxying the file through the API service and improves performance.

## OpenTelemetry Integration

The API service is instrumented with OpenTelemetry to provide tracing and metrics.
The service uses the OpenTelemetry Collector to export telemetry data to a backend of choice (e.g., Jaeger, Prometheus).

Tracing is used to track requests as they flow through the system, from the initial HTTP request to the database queries and other service calls.

## Data Models

This section provides a detailed description of the Go structs used in the API and their validation rules.

### Project

Represents a user project.

```go
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
```

**Validation Rules:**

-   `Name`: Must be between 1 and 100 characters.

### Image

Represents an image within a project.

```go
type Image struct {
	ID          string       `json:"id"`
	ProjectID   string       `json:"project_id"`
	OriginalURL string       `json:"original_url"`
	StagedURL   string       `json:"staged_url,omitempty"`
	RoomType    string       `json:"room_type,omitempty"`
	Style       string       `json:"style,omitempty"`
	Seed        int64        `json:"seed,omitempty"`
	Status      ImageStatus  `json:"status"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
```

**Validation Rules:**

-   `ProjectID`: Must be a valid UUID.
-   `OriginalURL`: Must be a valid URL.

### PresignUploadRequest

Represents the request to generate a presigned URL for file upload.

```go
type PresignUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
}
```

**Validation Rules:**

-   `Filename`: Must be between 1 and 255 characters.
-   `ContentType`: Must be one of `image/jpeg`, `image/png`, or `image/webp`.
-   `FileSize`: Must be between 1 and 10485760 bytes (10MB).

## Error Handling

The API service uses a consistent error handling strategy to provide clear and informative error messages to the client.

### Common Error Responses

-   **400 Bad Request:** The request could not be understood by the server due to malformed syntax.
-   **401 Unauthorized:** The request requires user authentication.
-   **404 Not Found:** The requested resource could not be found.
-   **422 Unprocessable Entity:** The request was well-formed but was unable to be followed due to semantic errors.
-   **500 Internal Server Error:** The server encountered an unexpected condition that prevented it from fulfilling the request.

### Error Response Structs

The API returns error responses in a consistent JSON format.

**`ErrorResponse`**

Used for general errors.

```go
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
```

**`ValidationErrorResponse`**

Used for validation errors.

```go
type ValidationErrorResponse struct {
	Error            string                  `json:"error"`
	Message          string                  `json:"message"`
	ValidationErrors []ValidationErrorDetail `json:"validation_errors"`
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
```
