# Configuration

This directory contains environment-specific configuration files for the Virtual Staging AI backend services (API and Worker).

## File Structure

- **`shared.yml`**: Base configuration shared across all environments. Contains default values.
- **`local.yml`**: Local development configuration (when running services locally, outside Docker)
- **`dev.yml`**: Docker Compose development environment configuration
- **`test.yml`**: Test environment configuration (used during integration tests)
- **`prod.yml`**: Production environment configuration (uses environment variable placeholders)

## How It Works

The configuration system uses a layered approach:

1. **`shared.yml`** is loaded first as the base configuration
2. The environment-specific file (e.g., `dev.yml`) is loaded next and overrides shared values
3. **`apps/<app>/secrets.yml`** is loaded next (if present) for app-specific secrets
4. Environment variables take final precedence and override YAML values

### Environment Selection

Set the `APP_ENV` environment variable to select which configuration to load:
- `APP_ENV=local` → loads `shared.yml` + `local.yml`
- `APP_ENV=dev` → loads `shared.yml` + `dev.yml` (default)
- `APP_ENV=test` → loads `shared.yml` + `test.yml`
- `APP_ENV=prod` → loads `shared.yml` + `prod.yml`

## Configuration Sections

### `app`
Application-level settings:
- `env`: Environment name (dev, test, prod, local)

### `auth0`
Auth0 authentication settings (API only):
- `audience`: Auth0 API audience
- `domain`: Auth0 domain

### `db`
PostgreSQL database configuration:
- `pgdatabase`: Database name
- `pghost`: Database host
- `pgpassword`: Database password
- `pgport`: Database port (default: 5432)
- `pguser`: Database user
- `pgsslmode`: SSL mode (disable, require, etc.)

You can also set `DATABASE_URL` as an environment variable to override individual settings.

### `job`
Job queue configuration:
- `queue_name`: Redis queue name (default: "default")
- `worker_concurrency`: Number of concurrent workers (default: 5)

### `logging`
Logging configuration:
- `level`: Log level (debug, info, warn, error)

### `otel`
OpenTelemetry configuration:
- `exporter_otlp_endpoint`: OTLP endpoint for traces (e.g., http://localhost:4318)
### `redis`
Redis configuration:
- `addr`: Redis address (e.g., localhost:6379)

### `replicate`
Replicate AI API configuration (Worker only):
- `api_token`: Replicate API token (should be set in `apps/worker/secrets.yml` or `REPLICATE_API_TOKEN` env var)
- `model_version`: Model version to use (default: qwen/qwen-image-edit)

### `s3`
S3/MinIO configuration:
- `access_key`: S3 access key
- `endpoint`: S3 endpoint URL (for MinIO/LocalStack)
- `public_endpoint`: Public S3 endpoint URL (for presigned URLs)
- `region`: AWS region (default: us-west-1)
- `secret_key`: S3 secret key
- `use_path_style`: Use path-style URLs (true for MinIO/LocalStack)

## Usage in Code

### API Service

```go
import "github.com/virtual-staging-ai/api/internal/config"

cfg, err := config.NewDefaultConfig()
if err != nil {
    log.Fatalf("failed to load config: %v", err)
}

// Access configuration
dbHost := cfg.GetPGHost()
s3Bucket := cfg.GetS3BucketName()
```

### Worker Service

```go
import "github.com/virtual-staging-ai/worker/internal/config"

cfg, err := config.NewDefaultConfig()
if err != nil {
    log.Fatalf("failed to load config: %v", err)
}

// Access configuration
replicateToken := cfg.GetReplicateAPIToken()
queueName := cfg.GetJobQueueName()
```

## Secrets Management

Sensitive values like API tokens should be stored in app-specific `secrets.yml` files:

- **`apps/api/secrets.yml`** - API service secrets (currently none required)
- **`apps/worker/secrets.yml`** - Worker service secrets (e.g., `REPLICATE_API_TOKEN`)

These files are:
- **HashiCorp Vault encrypted** - Always encrypted at rest using Vault
- **Gitignored** - Never committed to version control (encrypted or not)
- **Optional** - Services work without them if secrets are provided via environment variables
- **Example files provided** - Copy `secrets.yml.example` to `secrets.yml` and fill in your values

**Note:** `secrets.yml` files must be decrypted before the application starts, or you can provide secrets via environment variables which override file-based configuration.

### Worker secrets.yml example:
```yaml
replicate:
  api_token: your_replicate_api_token_here
```

## Docker Compose

The `docker-compose.yml` file mounts both config and secrets as read-only volumes:

```yaml
services:
  worker:
    environment:
      - APP_ENV=dev
    volumes:
      - ./config:/config:ro
      - ./apps/worker/secrets.yml:/app/secrets.yml:ro
```

Only `APP_ENV` is set as an environment variable. All other configuration is loaded from YAML files.

## Production Deployment

In production:

1. Set `APP_ENV=prod`
2. Override sensitive values using environment variables:
   ```bash
   export AUTH0_AUDIENCE=https://api.production.example.com
   export AUTH0_DOMAIN=production.us.auth0.com
   export DATABASE_URL=postgres://user:pass@host:5432/db
   export REPLICATE_API_TOKEN=your_token
   export S3_ACCESS_KEY=your_key
   export S3_SECRET_KEY=your_secret
   ```

The `prod.yml` file uses environment variable placeholders (e.g., `${AUTH0_AUDIENCE}`) that will be populated from the environment.

## Web Application

The web application (Next.js) continues to use environment variables directly via `.env.local` and `process.env.*`. See `apps/web/env.example` for required variables.

## Security Notes

- **Never commit secrets** to these YAML files
- Use environment variables for sensitive values (tokens, passwords, keys)
- The `config/` directory should be mounted read-only in containers
- In production, use secret management solutions (AWS Secrets Manager, HashiCorp Vault, etc.)
