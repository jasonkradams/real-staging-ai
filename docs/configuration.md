# Configuration

This document provides a detailed explanation of all the environment variables used in the Virtual Staging AI project.

## API Service (`api`)

| Variable                      | Description                                                                                                                                           | Default Value                      |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| `DATABASE_URL`                | Full Postgres DSN. If set, takes precedence over PG\* vars below. Format: `postgres://user:pass@host:port/db?sslmode=disable`                         |                                    |
| `APP_ENV`                     | The application environment.                                                                                                                          | `dev`                              |
| `AUTH0_DOMAIN`                | Your Auth0 domain.                                                                                                                                    |                                    |
| `AUTH0_AUDIENCE`              | The audience for your Auth0 API.                                                                                                                      | `https://api.virtualstaging.local` |
| `PGHOST`                      | The hostname of the PostgreSQL database.                                                                                                              | `postgres`                         |
| `PGPORT`                      | The port of the PostgreSQL database.                                                                                                                  | `5432`                             |
| `PGUSER`                      | The username for the PostgreSQL database.                                                                                                             | `postgres`                         |
| `PGPASSWORD`                  | The password for the PostgreSQL database.                                                                                                             | `postgres`                         |
| `PGDATABASE`                  | The name of the PostgreSQL database.                                                                                                                  | `virtualstaging`                   |
| `PGSSLMODE`                   | Postgres SSL mode when constructing DSN from PG\* vars.                                                                                               | `disable`                          |
| `REDIS_ADDR`                  | The address of the Redis server.                                                                                                                      | `redis:6379`                       |
| `JOB_QUEUE_NAME`              | Default Asynq queue name used by the API enqueuer.                                                                                                    | `default`                          |
| `STRIPE_WEBHOOK_SECRET`       | Required in non-dev environments; used to verify Stripe webhooks. **CRITICAL for production security - webhook verification will fail without this.** |                                    |
| `STRIPE_PUBLISHABLE_KEY`      | Stripe publishable key for frontend integration.                                                                                                      |                                    |
| `STRIPE_SECRET_KEY`           | Stripe secret key for server-side operations. **CRITICAL for production - required for payment processing.**                                          |                                    |
| `S3_ENDPOINT`                 | The endpoint of the S3-compatible storage.                                                                                                            | `http://minio:9000`                |
| `S3_PUBLIC_ENDPOINT`          | Public/base endpoint to use when presigning URLs (ensures browser-accessible host); when set, presigners use this host.                               |                                    |
| `S3_REGION`                   | The region of the S3 bucket.                                                                                                                          | `us-west-1`                        |
| `S3_BUCKET`                   | The name of the S3 bucket.                                                                                                                            | `virtual-staging`                  |
| `S3_BUCKET_NAME`              | Alias for S3 bucket name used in some code paths (fallback if `S3_BUCKET` not set).                                                                   |                                    |
| `S3_ACCESS_KEY`               | The access key for the S3 bucket.                                                                                                                     | `minioadmin`                       |
| `S3_SECRET_KEY`               | The secret key for the S3 bucket.                                                                                                                     | `minioadmin`                       |
| `S3_USE_PATH_STYLE`           | Whether to use path-style addressing for S3.                                                                                                          | `true`                             |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | The endpoint of the OpenTelemetry Collector.                                                                                                          | `http://otel:4318`                 |

## Worker Service (`worker`)

| Variable                      | Description                                  | Default Value       |
| ----------------------------- | -------------------------------------------- | ------------------- |
| `APP_ENV`                     | The application environment.                 | `dev`               |
| `PGHOST`                      | The hostname of the PostgreSQL database.     | `postgres`          |
| `PGPORT`                      | The port of the PostgreSQL database.         | `5432`              |
| `PGUSER`                      | The username for the PostgreSQL database.    | `postgres`          |
| `PGPASSWORD`                  | The password for the PostgreSQL database.    | `postgres`          |
| `PGDATABASE`                  | The name of the PostgreSQL database.         | `virtualstaging`    |
| `REDIS_ADDR`                  | The address of the Redis server.             | `redis:6379`        |
| `JOB_QUEUE_NAME`              | Default Asynq queue name to listen on.       | `default`           |
| `WORKER_CONCURRENCY`          | Number of concurrent workers.                | `5`                 |
| `S3_ENDPOINT`                 | The endpoint of the S3-compatible storage.   | `http://minio:9000` |
| `S3_REGION`                   | The region of the S3 bucket.                 | `us-west-1`         |
| `S3_BUCKET`                   | The name of the S3 bucket.                   | `virtual-staging`   |
| `S3_ACCESS_KEY`               | The access key for the S3 bucket.            | `minioadmin`        |
| `S3_SECRET_KEY`               | The secret key for the S3 bucket.            | `minioadmin`        |
| `S3_USE_PATH_STYLE`           | Whether to use path-style addressing for S3. | `true`              |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | The endpoint of the OpenTelemetry Collector. | `http://otel:4318`  |

## Security Notes

- Stripe Webhooks
  - In non-dev environments, `STRIPE_WEBHOOK_SECRET` is required. The API will fail closed (HTTP 503) if it is missing.
  - Webhook verification uses HMAC-SHA256 of `t.payload` with a timestamp tolerance (default 5m). Requests with invalid signatures or timestamps outside the tolerance are rejected (HTTP 401).
  - Rotation guidance:
    - Generate a new webhook secret in Stripe Dashboard.
    - Deploy the new secret as a platform secret/variable (e.g., GitHub Actions, Docker secrets, or cloud secret manager).
    - Roll out to all environments. Verify by sending a test event from Stripe CLI.
    - Remove the old secret once traffic is confirmed on the new key.
