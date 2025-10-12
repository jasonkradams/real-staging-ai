### Overview
A DIY virtual staging SaaS for real-estate photos. Upload a room photo → receive an AI-staged image using Replicate's SDXL-Lightning model. Fast time-to-market with production-ready AI staging.

### Tech Summary
- **Backend/API**: Go (Echo), Postgres (pgx), Redis (asynq), S3, Stripe, Auth0 (OIDC/JWT), OpenTelemetry
- **Frontend**: Next.js + Tailwind + shadcn/ui
- **AI Staging**: Replicate API (qwen/qwen-image-edit) - ~9s per image, ~$0.011/image
- **Infra**: Docker Compose (dev), GitHub Actions (CI), Fly.io/Render/Neon/Supabase/Cloudflare R2 (later)

### How It Works
1. User logs in via Auth0 → frontend gets JWT
2. Upload original image(s) via **S3 presigned PUT** from API
   - Supports single or **multi-file upload** (up to 50 images per batch)
   - Set default room type/style, or customize per image
3. Create **image job(s)** → enqueued to Redis (asynq)
4. Worker processes jobs: downloads original from S3, sends to Replicate AI for staging, uploads result back to S3
5. API marks image **ready** → client fetches results / receives event updates via SSE

---

## Quickstart

1. Install dependencies: Docker, Docker Compose, Go 1.22+, Node.js 18+.

2. Get a Replicate API token:
   - Sign up at [replicate.com](https://replicate.com)
   - Get your token from [account settings](https://replicate.com/account/api-tokens)
   - Export it: `export REPLICATE_API_TOKEN=r8_your_token_here`

3. Start dev stack (API, Worker, Postgres, Redis, MinIO):

```bash
export REPLICATE_API_TOKEN=r8_your_token_here
make up
```

4. Open the web app at: http://localhost:3000

5. Open API docs at: http://localhost:8080/api/v1/docs/

6. Basic health check:

```bash
curl -s http://localhost:8080/health
```

> **Note**: See `docs/REPLICATE_SETUP.md` for detailed configuration and troubleshooting.

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

- **Replicate AI** (Required for staging)
  - `REPLICATE_API_TOKEN`: Your Replicate API token (required)
  - `REPLICATE_MODEL_VERSION`: Model to use (default: `qwen/qwen-image-edit`)

- **Queue**
  - `REDIS_ADDR`: Redis address (required for job queue and SSE).
  - `JOB_QUEUE_NAME`: Asynq queue name (default `default`).
  - `WORKER_CONCURRENCY`: Worker concurrency (default `5`).

- **S3**
  - `S3_BUCKET_NAME`: S3 bucket name (required)
  - `S3_ENDPOINT`: Custom S3 endpoint (e.g., MinIO for local dev)
  - Local dev uses MinIO via `docker-compose.yml`

- **Stripe Webhooks**
  - `STRIPE_WEBHOOK_SECRET` (required in non-dev): verified with HMAC-SHA256 and timestamp tolerance.

## Documentation

**Full Documentation:**
- **Live:** https://jasonkradams.github.io/real-staging-ai/ (Material for MkDocs)
- **API Reference:** https://jasonkradams.github.io/real-staging-ai/api/ (OpenAPI/Swagger)
- **Local:** `make docs-serve` then open http://localhost:8000

**Local Development:**
```bash
# Serve MkDocs site
make docs-serve

# Validate OpenAPI spec
make docs-validate

# Build static site
make docs-build
```

## Conventional Commits

We use Conventional Commits. Examples:

- `feat(api): add presign upload endpoint`
- `fix(worker): handle empty staged_url gracefully`
- `docs(readme): expand quickstart and testing sections`

## Monorepo Structure

- `apps/api`: Go HTTP API (Echo), domain packages under `internal/<domain>`.
- `apps/worker`: Go background worker (Asynq for queue), publishes SSE via Redis.
- `apps/web`: Next.js frontend application.
- `apps/docs`: Material for MkDocs documentation site.
- `infra/migrations`: SQL migrations.
- `config/`: Environment-specific configuration files (dev, test, prod).
- `apps/api/web/api/v1`: OpenAPI spec and embedded Swagger UI.
