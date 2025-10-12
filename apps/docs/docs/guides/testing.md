# Testing Guide

Real Staging AI follows Test-Driven Development (TDD) principles with comprehensive unit and integration testing.

## Testing Philosophy

> **Rule #1: Tests First**  
> We write failing tests for each feature, then implement minimal code to make them pass.

### Core Principles

1. **Red-Green-Refactor**: Write failing test → Make it pass → Refactor
2. **High Coverage**: Target 80%+ for unit tests, 100% for critical paths
3. **Fast Feedback**: Unit tests run in seconds
4. **Real Dependencies**: Integration tests use actual PostgreSQL, Redis, S3

## Test Layers

### 1. Unit Tests

Fast, isolated tests with mocked dependencies.

**What We Test:**
- HTTP handlers
- Service layer logic
- Repository methods
- Input validation
- Business rules

**Tools:**
- Standard Go `testing` package
- Table-driven test pattern
- `storage.DatabaseMock` for database operations
- `pgxmock` for connection pooling tests

**Example:**
```go
func TestCreateProject_Success(t *testing.T) {
    t.Run("success: creates project with valid input", func(t *testing.T) {
        db := &storage.DatabaseMock{
            CreateProjectFunc: func(ctx context.Context, params CreateProjectParams) (Project, error) {
                return Project{
                    ID:     "proj_123",
                    Name:   params.Name,
                    UserID: params.UserID,
                }, nil
            },
        }
        
        service := NewProjectService(db)
        result, err := service.CreateProject(ctx, "Test Project", "user_123")
        
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if result.Name != "Test Project" {
            t.Error("project name mismatch")
        }
    })
    
    t.Run("fail: rejects empty project name", func(t *testing.T) {
        // Test validation...
    })
}
```

### 2. Integration Tests

End-to-end tests with real infrastructure via Docker Compose.

**What We Test:**
- Full HTTP request/response cycles
- Database transactions
- Redis queue operations
- S3 file operations
- Auth middleware
- Worker job processing

**Infrastructure:**
- PostgreSQL (docker-compose.test.yml)
- Redis
- LocalStack (S3-compatible)
- Migrations applied automatically

**Example:**
```go
//go:build integration
// +build integration

func TestIntegration_CreateImage(t *testing.T) {
    // Real database connection
    db, err := pgx.Connect(context.Background(), testDBURL)
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    
    // Make real HTTP request
    resp, err := http.Post(
        "http://localhost:8080/api/v1/images",
        "application/json",
        body,
    )
    // Assert response...
}
```

### 3. Contract Tests

Validate API contracts and external integrations.

**What We Test:**
- OpenAPI specification compliance
- Stripe webhook payload structure
- Auth0 JWT claims
- Replicate API responses

## Running Tests

### Unit Tests

```bash
# Run all unit tests
make test

# Run with coverage
make test-cover

# Run specific package
cd apps/api
go test -v ./internal/project

# Run specific test
go test -v -run TestCreateProject
```

**Output:**
```
Running unit tests...
--> Running api tests
ok      github.com/real-staging-ai/api/internal/project  0.045s
ok      github.com/real-staging-ai/api/internal/image    0.062s
--> Running worker tests
ok      github.com/real-staging-ai/worker/internal/staging  0.053s
```

### Integration Tests

```bash
# Run all integration tests
make test-integration

# This will:
# 1. Start docker-compose.test.yml
# 2. Run migrations
# 3. Execute integration tests
# 4. Clean up containers
```

**Optional E2E Upload Test:**

Test the complete presign → upload → process → ready flow:

```bash
# Run with E2E tests enabled
RUN_E2E_UPLOAD_READY=1 make test-integration
```

Or with explicit commands:
```bash
cd apps/api
CONFIG_DIR=../../config APP_ENV=test \
PGHOST=localhost PGPORT=5433 \
PGUSER=testuser PGPASSWORD=testpassword \
PGDATABASE=testdb PGSSLMODE=disable \
REDIS_ADDR=localhost:6379 \
RUN_E2E_UPLOAD_READY=1 \
go test -tags=integration -p 1 -v ./...
```

### Web Tests

```bash
# Run frontend tests
cd apps/web
npm run test

# With coverage
npm run test:coverage
```

## Test Organization

### File Naming

- Unit tests: `*_test.go` (same package)
- Integration tests: `*_test.go` with `//go:build integration` tag
- Mocks: `*_mock.go` (auto-generated)

### Test Naming Convention

Test functions follow a consistent pattern:

```go
func TestFunction_Scenario(t *testing.T) {
    t.Run("success: happy path description", func(t *testing.T) {
        // Test implementation
    })
    
    t.Run("fail: specific error condition", func(t *testing.T) {
        // Test implementation
    })
}
```

**Examples:**
- `TestCreateProject_Success`
- `TestCreateProject_EmptyName`
- `TestCreateProject_DatabaseError`
- `TestGetProject_NotFound`

### Table-Driven Tests

For testing multiple scenarios:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"success: valid email", "user@example.com", false},
        {"fail: missing @", "userexample.com", true},
        {"fail: empty string", "", true},
        {"fail: missing domain", "user@", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Test Coverage

### Viewing Coverage

```bash
# Generate coverage report
make test-cover

# Opens coverage.html in browser
# Shows line-by-line coverage

# Command-line coverage
go test -cover ./...
```

### Coverage Targets

| Package Type | Target |
|--------------|--------|
| Critical paths (auth, billing) | 100% |
| Business logic | 90%+ |
| HTTP handlers | 85%+ |
| Repository layer | 80%+ |
| Overall | 80%+ |

### Checking Coverage

```bash
# Coverage by package
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# HTML report
go tool cover -html=coverage.out -o coverage.html
```

## Example Test Scenarios

### Authentication Tests

```go
func TestJWTMiddleware_ValidToken(t *testing.T) {
    t.Run("success: accepts valid Auth0 JWT", func(t *testing.T) {
        // Mock JWKS endpoint
        // Create valid token
        // Make request with token
        // Assert 200 OK
    })
    
    t.Run("fail: rejects expired token", func(t *testing.T) {
        // Assert 401 Unauthorized
    })
    
    t.Run("fail: rejects missing token", func(t *testing.T) {
        // Assert 401 Unauthorized
    })
}
```

### API Endpoint Tests

```go
func TestPresignUpload_Success(t *testing.T) {
    t.Run("success: returns presigned URL and key", func(t *testing.T) {
        req := PresignUploadRequest{
            Filename:    "room.jpg",
            ContentType: "image/jpeg",
            FileSize:    1024000,
        }
        
        resp, err := handler.PresignUpload(ctx, req)
        
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if resp.UploadURL == "" {
            t.Error("upload URL is empty")
        }
        if resp.Key == "" {
            t.Error("S3 key is empty")
        }
    })
    
    t.Run("fail: rejects oversized file", func(t *testing.T) {
        // Test file size validation
    })
}
```

### Worker Tests

```go
func TestStageRun_Success(t *testing.T) {
    t.Run("success: processes image and updates status", func(t *testing.T) {
        // Setup: create test image in S3
        // Execute: process staging job
        // Assert: staged image uploaded
        // Assert: database status = ready
        // Assert: SSE event published
    })
    
    t.Run("fail: handles Replicate API error", func(t *testing.T) {
        // Mock Replicate error
        // Assert: status = error
        // Assert: error message set
    })
}
```

### Database Tests

```go
func TestCreateProject_Transaction(t *testing.T) {
    t.Run("success: commits transaction", func(t *testing.T) {
        // Begin transaction
        // Create project
        // Commit
        // Verify project exists
    })
    
    t.Run("fail: rolls back on error", func(t *testing.T) {
        // Begin transaction
        // Trigger error
        // Verify no project created
    })
}
```

### Stripe Webhook Tests

```go
func TestStripeWebhook_CheckoutCompleted(t *testing.T) {
    t.Run("success: handles checkout.session.completed", func(t *testing.T) {
        // Create webhook payload
        // Sign with test secret
        // POST to /api/v1/stripe/webhook
        // Assert user plan updated
        // Assert idempotency record created
    })
    
    t.Run("fail: rejects invalid signature", func(t *testing.T) {
        // Assert 400 Bad Request
    })
}
```

### SSE Tests

```go
func TestSSE_ImageReady(t *testing.T) {
    t.Run("success: receives image.ready event", func(t *testing.T) {
        // Connect to SSE endpoint
        // Trigger image ready event
        // Assert event received
        // Assert event data correct
    })
}
```

## Mocking

### Database Mocking

Use generated mocks:

```go
db := &storage.DatabaseMock{
    GetUserFunc: func(ctx context.Context, id string) (*User, error) {
        return &User{ID: id, Name: "Test User"}, nil
    },
}

// Verify mock was called
if len(db.GetUserCalls()) != 1 {
    t.Error("GetUser not called")
}
```

### Generating Mocks

```bash
# Generate all mocks
make generate

# This runs:
# - sqlc generate
# - moq for interface mocks
```

**Mock generation annotations:**

```go
//go:generate go run github.com/matryer/moq@v0.5.3 -out storage_mock.go . Database

type Database interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, params CreateUserParams) (*User, error)
}
```

## Test Utilities

### Test Fixtures

```go
// testdata/fixtures.go
func CreateTestUser(t *testing.T, db Database) *User {
    user, err := db.CreateUser(context.Background(), CreateUserParams{
        Auth0Sub: "auth0|test-" + uuid.NewString(),
        Role:     "user",
    })
    if err != nil {
        t.Fatalf("failed to create test user: %v", err)
    }
    return user
}
```

### Test Data

Place test files in `testdata/` directories:

```
apps/api/tests/integration/testdata/
├── seed.sql          # Database seed data
├── test-image.jpg    # Sample images
└── webhooks/
    └── stripe-checkout-completed.json
```

### Environment Setup

```go
func setupTestEnv(t *testing.T) {
    t.Setenv("APP_ENV", "test")
    t.Setenv("PGHOST", "localhost")
    t.Setenv("PGPORT", "5433")
    // More env vars...
}
```

## CI/CD Integration

### GitHub Actions

Tests run automatically on pull requests:

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Run unit tests
        run: make test
      
      - name: Run integration tests
        run: make test-integration
```

### Manual E2E Trigger

```bash
# Trigger E2E tests manually
gh workflow run CI \
  --field run_integration_tests=true \
  --field run_e2e_upload_ready=true
```

## Troubleshooting Tests

### Tests Fail with Database Connection Error

**Cause:** PostgreSQL not running or wrong connection string

**Solution:**
```bash
# Check database is running
docker compose -f docker-compose.test.yml ps postgres-test

# Verify connection
psql -h localhost -p 5433 -U testuser -d testdb
```

### Integration Tests Hang

**Cause:** Port conflicts or containers not stopping

**Solution:**
```bash
# Stop all test containers
docker compose -f docker-compose.test.yml down

# Clean up
docker compose -f docker-compose.test.yml down -v
```

### Mock Not Called

**Cause:** Mock function not set or logic not executed

**Solution:**
```go
// Verify mock setup
if db.GetUserFunc == nil {
    t.Fatal("GetUserFunc not set")
}

// Check call count
if len(db.GetUserCalls()) == 0 {
    t.Error("GetUser was not called")
}
```

## Best Practices

### DO:

✅ Write tests before implementation (TDD)  
✅ Test both success and failure cases  
✅ Use descriptive test names with `success:` and `fail:` prefixes  
✅ Keep tests focused and independent  
✅ Use table-driven tests for multiple scenarios  
✅ Mock external dependencies  
✅ Clean up test resources  
✅ Use t.Helper() for test helper functions  

### DON'T:

❌ Skip tests for "simple" code  
❌ Share state between tests  
❌ Use sleep for timing (use channels/context)  
❌ Ignore test failures  
❌ Test implementation details  
❌ Write overly complex test setups  
❌ Use global variables in tests  

## Learning Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [TestContainers](https://golang.testcontainers.org/)
- [Repository TDD Guide](../development/repository-guidelines.md)

---

**Related:**
- [Contributing Guide](../development/contributing.md)
- [Repository Guidelines](../development/repository-guidelines.md)
- [API Service](../architecture/api-service.md)
