> **Rule #1: Tests first.** We write failing tests for each feature, then implement minimal code to pass.

### Test Layers
1. **Unit tests** (Go): handlers, services, repos (use `sqlmock` or test DB), JWT middleware mocks, S3 presign logic
2. **Integration tests**: dockerized Postgres/Redis/minio (S3-compatible), run migrations, execute API flows end-to-end
3. **Contract tests**: API OpenAPI validation (oapi-codegen or kin-openapi), Stripe webhook payloads, Auth0 JWKS cache behavior

### Targets
- `make test` → unit tests (fast)
- `make test-integration` → spins docker-compose.test (pg, redis, minio) and runs end-to-end flows

### Example Test Scenarios (Phase 1)
- **Auth**: reject requests without/invalid JWT; accept with valid Auth0 JWT (mock JWKS)
- **Presign Upload**: POST `/v1/uploads/presign` returns URL, key; enforces content-type/size
- **Create Image Job**: POST `/v1/images` inserts rows, enqueues asynq, returns 202 w/ id
- **Get Image**: GET `/v1/images/{id}` returns status transitions: `queued` → `processing` → `ready`
- **Worker**: given a job, downloads original (minio), creates placeholder staged, uploads, updates DB
- **Stripe Webhook**: handles `checkout.session.completed`, sets plan & `stripe_customer_id`
- **SSE**: `/v1/events` streams job updates (can be tested with a short-lived server + client)

### Test Utilities
- Testcontainers or docker-compose for integration
- Seed data fixtures for users/projects
- Golden files for minimal image placeholder bytes
