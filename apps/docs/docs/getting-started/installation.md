# Installation

This guide walks through setting up Real Staging AI for local development.

## Prerequisites

### Required Software

Ensure you have the following installed:

| Tool | Version | Purpose |
|------|---------|---------|
| **Docker** | 20.10+ | Container runtime |
| **Docker Compose** | 2.0+ | Multi-container orchestration |
| **Go** | 1.22+ | Backend API and worker |
| **Node.js** | 18+ | Frontend application |
| **Make** | Any | Build automation |

### External Services

You'll need accounts for:

- **[Replicate](https://replicate.com)** - AI image processing (required)
- **[Auth0](https://auth0.com)** - Authentication (required for full flow)
- **[Stripe](https://stripe.com)** - Payment processing (optional for dev)

## Step 1: Clone Repository

```bash
git clone https://github.com/jasonkradams/real-staging-ai.git
cd real-staging-ai
```

## Step 2: Configure Replicate

Replicate provides the AI models for image staging.

1. Sign up at [replicate.com](https://replicate.com)
2. Navigate to [API Tokens](https://replicate.com/account/api-tokens)
3. Create a new token
4. Export it:

```bash
export REPLICATE_API_TOKEN=r8_your_token_here
```

!!! tip "Persistent Configuration"
    Add the export to your `~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish` to persist across sessions.

## Step 3: Configure Auth0 (Optional for Basic Testing)

For full authentication flow:

1. Create a free Auth0 account at [auth0.com](https://auth0.com)
2. Create a new Application (Single Page Application)
3. Configure settings:
   - **Allowed Callback URLs**: `http://localhost:3000/api/auth/callback`
   - **Allowed Logout URLs**: `http://localhost:3000`
   - **Allowed Web Origins**: `http://localhost:3000`
4. Copy configuration to `apps/api/secrets.yml` and `apps/worker/secrets.yml`

See the [Authentication Guide](../guides/authentication.md) for detailed setup.

## Step 4: Start Development Stack

The `make up` command starts all services via Docker Compose:

```bash
make up
```

This starts:

- **PostgreSQL** (port 5432) - Database
- **Redis** (port 6379) - Job queue and SSE
- **MinIO** (port 9000/9001) - S3-compatible storage
- **API** (port 8080) - Go HTTP API
- **Worker** - Background job processor
- **OpenTelemetry Collector** (port 4317/4318) - Observability
- **Web** (port 3000) - Next.js frontend

### Verify Services

Check that services are running:

```bash
# Health check
curl -s http://localhost:8080/health | jq

# Expected output:
# {
#   "status": "healthy",
#   "database": "connected",
#   "redis": "connected"
# }
```

Open the web app:
```
http://localhost:3000
```

Open the API docs:
```
http://localhost:8080/api/v1/docs/
```

## Step 5: Run Migrations

Database migrations run automatically with `make up`, but you can run them manually:

```bash
# Apply all migrations
make migrate

# Rollback all migrations
make migrate-down-dev
```

## Step 6: Verify Installation

### Run Tests

```bash
# Unit tests
make test

# Integration tests (starts test containers)
make test-integration
```

### Create Test Token

Generate an Auth0 token for API testing:

```bash
make token
```

Copy the token and use it in API requests:

```bash
export TOKEN=$(make token)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/projects
```

## Configuration Files

### Environment Structure

```
config/
├── local.yml    # Local development (default)
├── dev.yml      # Development environment
├── prod.yml     # Production environment
└── README.md
```

### Secrets Files

Create `apps/api/secrets.yml` and `apps/worker/secrets.yml` from examples:

```bash
cp apps/api/secrets.yml.example apps/api/secrets.yml
cp apps/worker/secrets.yml.example apps/worker/secrets.yml
```

Edit with your credentials. **Never commit secrets files!**

## Troubleshooting

### Docker Issues

**Port conflicts:**
```bash
# Check what's using a port
lsof -i :8080

# Stop all containers
make down
```

**Permission issues:**
```bash
# Fix Docker permissions (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

### Database Issues

**Connection refused:**
```bash
# Check PostgreSQL is healthy
docker compose ps postgres

# View logs
docker compose logs postgres
```

**Reset database:**
```bash
make migrate-down-dev
make migrate
```

### Replicate Issues

**Invalid token:**
```bash
# Verify token is set
echo $REPLICATE_API_TOKEN

# Test token
curl -H "Authorization: Token $REPLICATE_API_TOKEN" \
  https://api.replicate.com/v1/models
```

### MinIO Issues

**Cannot upload files:**

1. Open MinIO console: http://localhost:9001
2. Login: `minioadmin` / `minioadmin`
3. Verify bucket exists: `real-staging`
4. Check bucket policy allows uploads

## Next Steps

- [Create Your First Project](first-project.md)
- [Learn about Configuration](../guides/configuration.md)
- [Explore the Architecture](../architecture/)

## Development Tips

### Hot Reload

- **API/Worker**: Rebuild with `docker compose up --build api worker`
- **Web**: Runs in dev mode with hot reload automatically

### Database Access

```bash
# Connect to PostgreSQL
docker compose exec postgres psql -U postgres -d realstaging
```

### View Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f api
docker compose logs -f worker
```

### Clean Up

```bash
# Remove containers and volumes
make clean-all

# Remove generated files
make clean
```

---

Continue to [Your First Project →](first-project.md)
