> **Rule #1: Tests first.** We write failing tests for each feature, then implement minimal code to pass.

### Test Layers
1. **Unit tests** (Go): handlers, services, repos (use `sqlmock` or test DB), JWT middleware mocks, S3 presign logic
2. **Integration tests**: dockerized Postgres/Redis/minio (S3-compatible), run migrations, execute API flows end-to-end
3. **Contract tests**: API OpenAPI validation (oapi-codegen or kin-openapi), Stripe webhook payloads, Auth0 JWKS cache behavior

### Targets
- `make test` → unit tests (fast)
- `make test-integration` → spins docker-compose.test (pg, redis, minio) and runs end-to-end flows

Optional end-to-end upload → ready flow (SSE-verified) is gated by an env flag:

```
RUN_E2E_UPLOAD_READY=1 make test-integration
```

### CI: Running the E2E happy path (optional)

You can trigger the optional E2E presign → upload → create image → SSE flow in GitHub Actions via a manual run:

1) In GitHub, go to Actions → CI → Run workflow.
2) Set these inputs:
- run_integration_tests: true
- run_e2e_upload_ready: true

This sets `RUN_E2E_UPLOAD_READY=1` only for that run and executes `make test-integration`.

Alternatively, from the CLI using `gh`:

```bash
gh workflow run CI \
  --field run_integration_tests=true \
  --field run_e2e_upload_ready=true
```

Note: The integration workflow also runs on a nightly schedule without E2E unless explicitly enabled via the input above.

### Example Test Scenarios (Phase 1)
- **Auth**: reject requests without/invalid JWT; accept with valid Auth0 JWT (mock JWKS)
- **Presign Upload**: POST `/v1/uploads/presign` returns URL, key; enforces content-type/size
- **Create Image Job**: POST `/v1/images` inserts rows, enqueues asynq, returns 202 w/ id
- **Get Image**: GET `/v1/images/{id}` returns status transitions: `queued` → `processing` → `ready`
- **Worker**: given a job, downloads original (minio), creates placeholder staged, uploads, updates DB
- **Stripe Webhook**: handles `checkout.session.completed`, sets plan & `stripe_customer_id`
- **SSE**: `/v1/events` streams job updates (can be tested with a short-lived server + client)
  - Full presign → upload → create image → SSE ready path exists under `apps/api/tests/integration/e2e_upload_ready_test.go` and is enabled when `RUN_E2E_UPLOAD_READY=1`.

### Test Utilities
- Testcontainers or docker-compose for integration
- Seed data fixtures for users/projects
- Golden files for minimal image placeholder bytes
