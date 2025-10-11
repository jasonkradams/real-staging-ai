### Monorepo Layout

```
/real-staging
  /apps
    /web               # Next.js UI (upload, dashboard, checkout)
    /api               # Go Echo service (REST, auth, Stripe, jobs)
    /worker            # Go asynq worker (Phase 1 mock staging)
  /infra
    docker-compose.yml
    /migrations        # SQL migrations (golang-migrate)
  /docs                # (these files)
```

### Services
- **API (Go + Echo)**: REST endpoints, Auth0 JWT validation (JWKS), presigned URLs, job enqueue, Stripe webhooks, SSE for job events.
- **Worker (Go + asynq)**: Consumes `stage:run` jobs. Phase 1: downloads original from S3 (or reads via presigned GET), creates a stamped copy (e.g., `-staged.jpg`), uploads back to S3, notifies API.
- **Postgres**: core metadata & usage
- **Redis**: asynq queue & retries
- **S3**: object storage for uploads and outputs

### Data Model (Phase 1)
- `users(id, auth0_sub, stripe_customer_id, role, created_at)`
- `projects(id, user_id, name, created_at)`
- `images(id, project_id, original_url, staged_url, room_type, style, status, error, seed, created_at, updated_at)`
- `jobs(id, image_id, type, payload_json, status, error, created_at, started_at, finished_at)`
- `plans(id, code, price_id, monthly_limit)`
- `usage(user_id, month_yyyymm, count)` (VIEW or materialized via query)

### Key Libraries (Go)
- **Echo**: github.com/labstack/echo/v4
- **pgx**: github.com/jackc/pgx/v5
- **sqlc**: type-safe queries (generate from SQL)
- **golang-migrate**: migrations
- **asynq**: github.com/hibiken/asynq
- **aws-sdk-go-v2**: S3 presign/get/put
- **stripe-go/v76**: Stripe API
- **go-jose / jwks**: Auth0 JWT validation (via Echo middleware or custom)
- **otel**: go.opentelemetry.io/otel + otel-echo, otel-pgx, otel-http
