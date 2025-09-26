### Overview
A DIY virtual staging SaaS for real-estate photos. Upload a room photo → receive a *mocked* staged image in Phase 1 (no GPU yet). We’ll validate flows, pricing, and UX while keeping the image pipeline simple.

### Tech Summary
- **Backend/API**: Go (Echo), Postgres (pgx), Redis (asynq), S3, Stripe, Auth0 (OIDC/JWT), OpenTelemetry
- **Frontend**: Next.js + Tailwind + shadcn/ui
- **Infra**: Docker Compose (dev), GitHub Actions (CI), Fly.io/Render/Neon/Supabase/Cloudflare R2 (later)

### Phase 1 Goals
- End-to-end flow working: auth → upload → job → result placeholder → billing → usage tracking
- **Test-driven** from the start (unit + integration)
- Image generation is **stubbed**: returns a watermarked placeholder based on the uploaded image to exercise data flow, S3, and queueing.

1. User logs in via Auth0 → frontend gets JWT
2. Upload original image via **S3 presigned PUT** from API
3. Create an **image job** → enqueued to Redis (asynq)
4. Worker processes job (Phase 1: copy original to a `-staged.jpg` variant + watermark)
5. API marks image **ready** → client fetches results / receives event updates

---

## Quickstart

1. Install dependencies: Docker, Docker Compose, Go 1.22+.
2. Start dev stack (API, Worker, Postgres, Redis, MinIO):

```bash
make up
```

3. Open API docs at: http://localhost:8080/api/v1/docs/

4. Basic health check:

```bash
curl -s http://localhost:8080/health
```

## Development

- Run unit tests:

```bash
make test
```

- Run integration tests (dockerized Postgres/Redis/LocalStack):

```bash
make test-integration
```

## Integration Tests

- Standard suite lives under `apps/api/tests/integration/` and `apps/worker/tests/integration/`.
- The suite brings up Postgres, Redis, and LocalStack via `docker-compose.test.yml`.
- Makefile sets the proper env for tests (PG*, `REDIS_ADDR`).

### Optional E2E Happy Path (env-gated)

End-to-end test that performs: presign → PUT upload (LocalStack S3) → create image → SSE emits processing → ready → DB `staged_url` set.

```bash
cd apps/api
PGHOST=localhost PGPORT=5433 PGUSER=testuser PGPASSWORD=testpassword PGDATABASE=testdb PGSSLMODE=disable \
REDIS_ADDR=localhost:6379 RUN_E2E_UPLOAD_READY=1 \
go test -tags=integration -v ./tests/integration -run TestE2E_Presign_Upload_CreateImage_ReadyViaSSE
```

## Configuration

See `docs/configuration.md` for all environment variables. Highlights:

- Queue
  - `REDIS_ADDR`: Redis address (required for job queue and SSE).
  - `JOB_QUEUE_NAME`: Asynq queue name (default `default`).
  - `WORKER_CONCURRENCY`: Worker concurrency (default `5`).

- Stripe Webhooks
  - `STRIPE_WEBHOOK_SECRET` (required in non-dev): verified with HMAC-SHA256 and timestamp tolerance.

- S3
  - Local dev uses LocalStack (MinIO alternative) via `docker-compose.test.yml`.

## API Docs

- OpenAPI is embedded and served at `/api/v1/docs`.
- Validate spec:

```bash
make docs
```

## Conventional Commits

We use Conventional Commits. Examples:

- `feat(api): add presign upload endpoint`
- `fix(worker): handle empty staged_url gracefully`
- `docs(readme): expand quickstart and testing sections`

## Monorepo Structure

- `apps/api`: Go HTTP API (Echo), domain packages under `internal/<domain>`.
- `apps/worker`: Go background worker (Asynq for queue), publishes SSE via Redis.
- `infra/migrations`: SQL migrations.
- `apps/api/web/api/v1`: OpenAPI spec and docs (embedded).
- `docs/`: architecture, configuration, review notes, and TODOs.
