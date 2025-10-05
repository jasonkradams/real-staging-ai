# Configuration System Migration Guide

This document describes the migration from environment-variable-based configuration to the new YAML-based configuration system.

## What Changed

### Before
- All configuration was done via environment variables
- Each service required 15-20 environment variables to be set
- docker-compose.yml files were cluttered with environment settings
- No easy way to share configuration across environments
- Difficult to override settings locally

### After
- Configuration is stored in YAML files under `config/`
- Environment variables are only used for sensitive data and environment selection
- Docker Compose only needs `APP_ENV` to be set
- Shared configuration with environment-specific overrides
- Full test coverage for configuration loading

## Configuration Files

### Location
All configuration files are located in the `config/` directory at the project root:

```
config/
├── README.md          # Configuration documentation
├── shared.yml         # Base configuration for all environments
├── local.yml          # Local development (outside Docker)
├── dev.yml            # Docker Compose development
├── test.yml           # Test environment
└── prod.yml           # Production environment
```

### File Structure
Each YAML file contains these sections:
- `app`: Application settings (environment name)
- `auth0`: Auth0 configuration (API only)
- `db`: PostgreSQL database settings
- `job`: Job queue settings
- `logging`: Logging configuration
- `otel`: OpenTelemetry settings
- `redis`: Redis settings
- `replicate`: Replicate AI settings (worker only)
- `s3`: S3/MinIO settings

## Migration Steps

### For API Service

**Before:**
```go
bucketName := os.Getenv("S3_BUCKET")
if bucketName == "" {
    bucketName = "virtual-staging"
}
```

**After:**
```go
import "github.com/virtual-staging-ai/api/internal/config"

cfg, err := config.NewDefaultConfig()
if err != nil {
    log.Fatalf("failed to load config: %v", err)
}
bucketName := cfg.GetS3BucketName()
```

### For Worker Service

**Before:**
```go
replicateToken := os.Getenv("REPLICATE_API_TOKEN")
if replicateToken == "" {
    return nil, fmt.Errorf("REPLICATE_API_TOKEN required")
}
```

**After:**
```go
import "github.com/virtual-staging-ai/worker/internal/config"

cfg, err := config.NewDefaultConfig()
if err != nil {
    log.Fatalf("failed to load config: %v", err)
}
replicateToken := cfg.GetReplicateAPIToken()
```

### For Docker Compose

**Before:**
```yaml
api:
  environment:
    - APP_ENV=dev
    - AUTH0_AUDIENCE=https://api.virtualstaging.local
    - AUTH0_DOMAIN=dev-sleeping-pandas.us.auth0.com
    - PGDATABASE=virtualstaging
    - PGHOST=postgres
    # ... 15 more variables
```

**After:**
```yaml
api:
  environment:
    - APP_ENV=dev
    - REPLICATE_API_TOKEN=${REPLICATE_API_TOKEN}
  volumes:
    - ./config:/config:ro
```

## Environment Variable Precedence

The configuration system uses a layered approach:

1. **YAML files** (lowest precedence):
   - `config/shared.yml` is loaded first
   - `config/{env}.yml` is loaded next and overrides shared values

2. **Environment variables** (highest precedence):
   - Environment variables always override YAML values
   - Useful for secrets and production overrides

### Example

`config/shared.yml`:
```yaml
db:
  pghost: localhost
  pgport: 5432
```

`config/dev.yml`:
```yaml
db:
  pghost: postgres  # Overrides shared.yml
```

Environment variable:
```bash
export PGHOST=custom-host  # Overrides both YAML files
```

Result: `cfg.GetPGHost()` returns `"custom-host"`

## Testing

The configuration system includes comprehensive tests:

```bash
# Test API config
cd apps/api
go test -v ./internal/config/...

# Test Worker config
cd apps/worker
go test -v ./internal/config/...
```

Both test suites achieve 100% code coverage.

## Backward Compatibility

### Legacy Environment Variables

The following legacy environment variables are still supported:

- `DATABASE_URL` - Overrides individual DB settings
- `S3_BUCKET` - Alias for `S3_BUCKET_NAME`
- All other env vars work as before when set

### Migration Path

You can migrate gradually:
1. Start using YAML files for non-sensitive configuration
2. Continue using environment variables for secrets
3. Update code to use the config package over time

## Production Deployment

### Recommended Approach

1. Set `APP_ENV=prod`
2. Mount `config/` directory (read-only)
3. Override sensitive values with environment variables:

```bash
export APP_ENV=prod
export AUTH0_AUDIENCE=https://api.production.example.com
export AUTH0_DOMAIN=production.us.auth0.com
export DATABASE_URL=postgres://...
export REPLICATE_API_TOKEN=r8_xxx
export S3_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE
export S3_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

### Kubernetes Example

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  shared.yml: |
    # Base configuration
  prod.yml: |
    # Production overrides
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
data:
  replicate-api-token: <base64-encoded>
  s3-access-key: <base64-encoded>
---
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: api
        env:
        - name: APP_ENV
          value: "prod"
        - name: REPLICATE_API_TOKEN
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: replicate-api-token
        volumeMounts:
        - name: config
          mountPath: /config
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: app-config
```

## Benefits

1. **Centralized Configuration**: All settings in one place
2. **Environment-Specific**: Easy to maintain different environments
3. **Type-Safe**: Go interface with proper types (int, bool, string)
4. **Testable**: 100% test coverage with comprehensive test cases
5. **Documented**: Self-documenting YAML files with comments
6. **Secure**: Secrets still use environment variables
7. **Flexible**: Layer-based override system
8. **Docker-Friendly**: Mount config as read-only volume

## Troubleshooting

### Config file not found

**Error**: `failed to read shared config: no such file or directory`

**Solution**: Set `CONFIG_DIR` environment variable or ensure `config/` directory exists:
```bash
export CONFIG_DIR=/path/to/config
```

### Wrong environment loaded

**Error**: Loading dev config in production

**Solution**: Ensure `APP_ENV` is set correctly:
```bash
export APP_ENV=prod
```

### Environment variable not working

**Issue**: YAML value used instead of environment variable

**Solution**: Ensure environment variable is exported before starting the service:
```bash
export PGHOST=myhost
./api-server
```

### Boolean values not parsing

**Issue**: `use_path_style: false` not working

**Solution**: Ensure boolean values in YAML are lowercase (`true`/`false`, not `True`/`False`)

## References

- Configuration Package: `apps/api/internal/config/` and `apps/worker/internal/config/`
- Configuration Files: `config/`
- Configuration Documentation: `config/README.md`
- Docker Compose: `docker-compose.yml`
- Library Used: [cleanenv](https://github.com/ilyakaznacheev/cleanenv)
