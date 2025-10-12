# Technology Stack

Detailed overview of the technologies powering Real Staging AI.

## Backend

### Go 1.22+

**Why Go?**
- Excellent performance for I/O-bound workloads
- Native concurrency with goroutines
- Strong typing and compile-time safety
- Fast compilation and deployment
- Rich ecosystem for cloud services

**Key Packages:**
- `net/http` - HTTP server and client
- `context` - Request context and cancellation
- `encoding/json` - JSON serialization
- `crypto/*` - Cryptographic operations

### Echo Framework

**Why Echo?**
- High performance (faster than Gin, standard net/http)
- Middleware support
- Excellent routing
- Built-in validation
- Easy testing

**Features Used:**
- HTTP routing and handlers
- Middleware chain
- Request binding and validation
- Context management
- Error handling

**Example:**
```go
e := echo.New()
e.Use(middleware.Logger())
e.Use(middleware.Recover())
e.GET("/health", healthHandler)
e.POST("/api/v1/projects", createProject, jwtMiddleware)
```

### PostgreSQL 17

**Why PostgreSQL?**
- ACID compliance for financial data
- JSON/JSONB support for flexible schemas
- Excellent query performance
- Mature ecosystem and tooling
- Strong data integrity

**Features Used:**
- JSONB columns for metadata
- UUIDs as primary keys
- Foreign key constraints
- Partial indexes
- Row-level security patterns

**Connection:** pgx (fastest Go PostgreSQL driver)

### Redis 8.2

**Why Redis?**
- In-memory speed
- Pub/Sub for real-time features
- Reliable job queue (Asynq)
- Simple data structures

**Use Cases:**
- Job queue (via Asynq)
- SSE pub/sub
- Future: caching, rate limiting

### MinIO / S3

**Why S3?**
- Industry standard object storage
- Scalable and durable
- Presigned URLs for direct uploads
- Cost-effective

**MinIO for Development:**
- S3-compatible local testing
- No cloud dependencies
- Fast iteration

**Structure:**
```
bucket/
├── uploads/    # Original images
└── staged/     # Processed images
```

## Frontend

### Next.js 14

**Why Next.js?**
- React with server-side rendering
- Excellent performance
- Built-in routing
- API routes
- Image optimization

**Features Used:**
- App router (new paradigm)
- Server components
- Client components for interactivity
- API routes for Auth0 callbacks
- Middleware for auth

### TypeScript

**Why TypeScript?**
- Type safety
- Better IDE support
- Fewer runtime errors
- Self-documenting code

### Tailwind CSS

**Why Tailwind?**
- Utility-first approach
- Fast development
- Consistent design system
- Small production bundle
- Dark mode support

**Custom Configuration:**
```js
theme: {
  extend: {
    colors: {
      primary: { /* blue-indigo gradient */ }
    }
  }
}
```

### shadcn/ui

**Why shadcn/ui?**
- Accessible components
- Customizable
- No runtime dependency
- Copy-paste components
- Built on Radix UI

**Components Used:**
- Button
- Card
- Dialog
- Form inputs
- Toast notifications

## AI & Processing

### Replicate

**Why Replicate?**
- No infrastructure management
- Pay-per-use pricing
- Fast inference
- Multiple model options
- Simple API

**Models:**
- `qwen/qwen-image-edit` - Fast, cost-effective
- `black-forest-labs/flux-kontext-max` - High quality

**Pricing:** ~$0.011 per image, ~9s processing time

### Asynq

**Why Asynq?**
- Redis-backed reliability
- Automatic retries
- Priority queues
- Scheduled tasks
- Web UI for monitoring

**Features:**
- Exponential backoff
- Dead letter queue
- Task uniqueness
- Graceful shutdown

## Authentication

### Auth0

**Why Auth0?**
- Industry-leading auth platform
- OAuth 2.0 / OIDC standard
- Managed service
- Social login support
- MFA built-in

**Features Used:**
- Single Page Application flow
- JWT with RS256
- User management
- Custom claims

## Payment Processing

### Stripe

**Why Stripe?**
- Industry leader
- Excellent API and docs
- Managed billing
- Compliance built-in
- Webhooks for automation

**Features Used:**
- Subscriptions
- One-time payments
- Webhooks
- Customer portal
- Invoice management

## Observability

### OpenTelemetry

**Why OpenTelemetry?**
- Vendor-neutral
- Industry standard
- Traces, metrics, logs
- Rich instrumentation

**Exporters:**
- Jaeger (traces)
- Prometheus (metrics)
- Loki (logs)

## Infrastructure

### Docker

**Why Docker?**
- Consistent environments
- Easy deployment
- Service isolation
- Compose for local dev

**Images:**
- Go apps: scratch-based (tiny)
- PostgreSQL: official image
- Redis: official image
- MinIO: official image

### Docker Compose

**Why Compose?**
- Multi-container orchestration
- Simple configuration
- Perfect for local development
- Health checks
- Dependency management

### GitHub Actions

**Why GitHub Actions?**
- Native to GitHub
- Free for open source
- Matrix builds
- Docker support
- Easy secrets management

**Workflows:**
- CI: test, lint, build
- Pages: deploy docs
- Scheduled integration tests

## Development Tools

### sqlc

**Why sqlc?**
- Type-safe SQL
- Compile-time validation
- No reflection overhead
- Works with existing SQL

**Example:**
```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;
```

Generates:
```go
func (q *Queries) GetUser(ctx context.Context, id string) (User, error)
```

### moq

**Why moq?**
- Simple interface mocking
- Minimal generated code
- No magic
- Easy to understand

**Usage:**
```go
//go:generate go run github.com/matryer/moq -out mock.go . Database

type Database interface {
    GetUser(ctx context.Context, id string) (*User, error)
}
```

### golangci-lint

**Why golangci-lint?**
- Fast linting
- 50+ linters in one tool
- Configurable
- CI-friendly

**Enabled Linters:**
- gofmt, goimports
- govet, errcheck
- staticcheck
- gosec (security)
- Many more...

### Material for MkDocs

**Why Material for MkDocs?**
- Beautiful documentation
- Easy customization
- Search built-in
- Mobile responsive
- Dark mode

## Technology Decisions

### Why Not Alternatives?

**Why not Node.js/Python for backend?**
- Go offers better performance
- Lower memory footprint
- Better concurrency model
- Simpler deployment (single binary)

**Why not GraphQL?**
- REST is simpler
- Better caching
- Easier debugging
- OpenAPI specification

**Why not MongoDB?**
- PostgreSQL handles JSON well
- Need ACID for billing
- Better query optimization
- Stronger consistency

**Why not Kubernetes (for now)?**
- Docker Compose sufficient for MVP
- Lower operational complexity
- Can migrate later if needed

## Version Requirements

| Technology | Minimum Version | Recommended |
|------------|----------------|-------------|
| Go | 1.22 | 1.22+ |
| Node.js | 18 | 20+ |
| PostgreSQL | 15 | 17 |
| Redis | 7 | 8+ |
| Docker | 20.10 | Latest |
| Docker Compose | 2.0 | Latest |

## Production Recommendations

### Managed Services

**Database:**
- AWS RDS (PostgreSQL)
- Neon
- Supabase

**Redis:**
- AWS ElastiCache
- Upstash
- Redis Cloud

**S3:**
- AWS S3
- Cloudflare R2
- Backblaze B2

**Hosting:**
- Fly.io (easiest)
- AWS ECS
- Google Cloud Run
- Render

## Performance Characteristics

### API Latency

| Operation | P50 | P95 | P99 |
|-----------|-----|-----|-----|
| GET /projects | 15ms | 30ms | 50ms |
| POST /images | 25ms | 50ms | 80ms |
| Presign upload | 20ms | 40ms | 60ms |

### Throughput

- API: 1000+ requests/second (single instance)
- Worker: 20 jobs/second (5 workers)
- Database: 5000+ queries/second

### Resource Usage

**API (per instance):**
- Memory: 128-256MB
- CPU: 0.5-1 core
- Disk: 100MB

**Worker (per instance):**
- Memory: 256-512MB
- CPU: 1-2 cores
- Disk: 1GB (image caching)

## Future Considerations

**Potential Additions:**
- **Kafka** - High-throughput event streaming
- **Elasticsearch** - Full-text search
- **Redis Cluster** - Horizontal scaling
- **CDN** - Static asset delivery (Cloudflare)
- **Message Queue** - RabbitMQ or SQS for complex workflows

---

**Related:**
- [Architecture Overview](index.md)
- [Deployment Guide](../operations/deployment.md)
- [Configuration](../guides/configuration.md)
