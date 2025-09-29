### Prereqs
- Go 1.22+
- Node 20+
- Docker/Docker Compose
- Auth0 app (SPA + API) → Domain, Audience
- Stripe keys
- AWS S3 bucket (or MinIO locally)

### Environment Variables

```sh
# API
PORT=8080
APP_ENV=dev
JWT_AUDIENCE=https://api.virtualstaging.local
JWT_ISSUER=https://YOUR_DOMAIN.auth0.com/

# Postgres
PGHOST=localhost
PGPORT=5432
PGUSER=postgres
PGPASSWORD=postgres
PGDATABASE=virtualstaging

# Redis
REDIS_ADDR=localhost:6379

# S3 / MinIO
S3_ENDPOINT=http://localhost:9000
S3_REGION=us-west-1
S3_BUCKET=virtual-staging
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_USE_PATH_STYLE=true

# Stripe
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...

# OTEL
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=virtual-staging-api
```
