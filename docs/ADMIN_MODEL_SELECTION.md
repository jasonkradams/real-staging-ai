# Admin Model Selection Feature

**Status**: ✅ Implemented  
**Version**: Phase 1 Complete

## Overview

The Admin Model Selection feature allows administrators to dynamically configure which AI model is used for virtual staging operations through a web UI and REST API.

## Architecture

### Database Schema

**Table**: `settings`
```sql
CREATE TABLE settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);
```

**Default Setting**:
- `active_model` = `"black-forest-labs/flux-kontext-max"`

### Backend Components

#### 1. Settings Domain Package
Location: `apps/api/internal/settings/`

**Files**:
- `model.go` - Data models (Setting, ModelInfo, UpdateSettingRequest)
- `repository.go` - Repository interface
- `default_repository.go` - PostgreSQL implementation
- `service.go` - Service interface
- `default_service.go` - Business logic implementation
- `*_mock.go` - Generated mocks for testing

**Key Features**:
- CRUD operations for settings
- Model validation before updates
- Hardcoded model registry (matches worker models)
- User tracking for audit trail

#### 2. Admin HTTP Handler
Location: `apps/api/internal/http/admin_handler.go`

**Endpoints**:
```
GET    /api/v1/admin/models          - List all available models
GET    /api/v1/admin/models/active   - Get currently active model
PUT    /api/v1/admin/models/active   - Update active model
GET    /api/v1/admin/settings        - List all settings
GET    /api/v1/admin/settings/:key   - Get specific setting
PUT    /api/v1/admin/settings/:key   - Update specific setting
```

**Authentication**: Requires JWT token (Auth0)

#### 3. Frontend Admin UI
Location: `apps/web/app/admin/settings/page.tsx`

**Features**:
- Visual model cards showing name, description, version
- Active model highlighted with badge
- One-click model activation
- Real-time success/error feedback
- Loading states and error handling
- Responsive design using shadcn/ui components

## API Reference

### List Models

**Request**:
```http
GET /api/v1/admin/models
Authorization: Bearer <token>
```

**Response**:
```json
{
  "models": [
    {
      "id": "qwen/qwen-image-edit",
      "name": "Qwen Image Edit",
      "description": "Fast image editing model optimized for virtual staging. Requires input image.",
      "version": "v1",
      "is_active": false
    },
    {
      "id": "black-forest-labs/flux-kontext-max",
      "name": "Flux Kontext Max",
      "description": "High-quality image generation and editing with advanced context understanding. Supports both text-to-image and image-to-image.",
      "version": "v1",
      "is_active": true
    }
  ]
}
```

### Get Active Model

**Request**:
```http
GET /api/v1/admin/models/active
Authorization: Bearer <token>
```

**Response**:
```json
{
  "id": "black-forest-labs/flux-kontext-max",
  "name": "Flux Kontext Max",
  "description": "High-quality image generation and editing with advanced context understanding. Supports both text-to-image and image-to-image.",
  "version": "v1",
  "is_active": true
}
```

### Update Active Model

**Request**:
```http
PUT /api/v1/admin/models/active
Authorization: Bearer <token>
Content-Type: application/json

{
  "value": "qwen/qwen-image-edit"
}
```

**Response**:
```json
{
  "message": "Active model updated successfully",
  "model_id": "qwen/qwen-image-edit"
}
```

**Validation**:
- Model ID must exist in the available models list
- Returns 400 if model ID is invalid

## Usage

### Accessing the Admin UI

1. Navigate to `/admin/settings` in your browser
2. Ensure you're authenticated with Auth0
3. View all available AI models
4. Click "Activate" on any model to make it active
5. Confirmation will appear when successful

### Programmatic Access

```typescript
// Fetch available models
const token = await getAccessTokenSilently();
const response = await fetch('/api/v1/admin/models', {
  headers: { Authorization: `Bearer ${token}` }
});
const { models } = await response.json();

// Update active model
await fetch('/api/v1/admin/models/active', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
  },
  body: JSON.stringify({ value: 'qwen/qwen-image-edit' }),
});
```

## Integration with Worker

The worker service will need to be updated to read the active model from the database instead of using a hardcoded value:

**Current** (Hardcoded):
```go
stagingCfg := &staging.ServiceConfig{
    ModelID: model.ModelFluxKontextMax, // Hardcoded
    // ...
}
```

**Future** (Dynamic - Phase 2):
```go
// Read from database
activeModel, err := settingsRepo.GetByKey(ctx, "active_model")
if err != nil {
    log.Warn(ctx, "failed to get active model, using default")
    activeModel = model.ModelFluxKontextMax
}

stagingCfg := &staging.ServiceConfig{
    ModelID: model.ModelID(activeModel.Value),
    // ...
}
```

## Testing

### Unit Tests
```bash
cd apps/api
go test -v ./internal/settings
```

**Coverage**: All service methods tested with mocks

### Integration Tests
```bash
# Start dev environment
make up

# Test via curl
TOKEN=$(make token)

# List models
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/admin/models

# Update active model
curl -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"value":"qwen/qwen-image-edit"}' \
  http://localhost:8080/api/v1/admin/models/active
```

### UI Testing
1. Start dev environment: `make up`
2. Navigate to `http://localhost:3000/admin/settings`
3. Login with Auth0
4. Test model switching functionality

## Security

### Authentication & Authorization
- All admin endpoints require valid JWT token
- JWT validated via Auth0 middleware
- User ID extracted from token for audit trail

### Future Enhancements (Phase 2)
- Role-based access control (RBAC)
- Admin-only role requirement
- Permission checks before allowing updates
- Audit log for all model changes

## Migration

### Applying the Migration

**Development**:
```bash
make migrate
```

**Test**:
```bash
make migrate-test
```

**Production**:
```bash
# Using migrate CLI
migrate -path ./infra/migrations \
  -database "postgres://user:pass@host:5432/db?sslmode=require" \
  up
```

### Rollback

```bash
# Development
make migrate-down-dev

# Or specific migration
migrate -path ./infra/migrations \
  -database "postgres://..." \
  down 1
```

## Monitoring

### Metrics to Track
- Model change frequency
- Active model distribution over time
- Failed model updates
- User who made changes

### Logging
All model updates are logged with:
- User ID
- Timestamp
- Old model ID
- New model ID

Example log:
```json
{
  "time": "2025-10-08T10:30:00Z",
  "level": "INFO",
  "msg": "active model updated",
  "model_id": "qwen/qwen-image-edit",
  "user_id": "auth0|123456"
}
```

## Known Limitations

### Phase 1 Constraints
1. **Static Worker Configuration**: Worker still uses hardcoded model; database value not read dynamically
2. **No Hot Reload**: Requires worker restart to pick up model changes
3. **No RBAC**: Any authenticated user can access admin endpoints
4. **No Audit History**: Only stores last update, not full history

### Planned Improvements (Phase 2)
- Worker reads model from database on startup
- Worker polls for model changes (or uses cache with TTL)
- Role-based access control
- Full audit history table
- Per-project model overrides

## Troubleshooting

### "Failed to fetch models"
- **Cause**: API server not running or authentication failed
- **Fix**: Ensure `make up` is running and you're logged in

### "Invalid model ID" error
- **Cause**: Trying to activate a model not in registry
- **Fix**: Only use model IDs from the available models list

### Worker using wrong model
- **Cause**: Phase 1 doesn't sync with worker yet
- **Fix**: Manually update `apps/worker/main.go` and restart worker

### Migration fails
- **Cause**: Database schema conflict
- **Fix**: Check existing schema, may need to rollback and reapply

## Related Documentation

- [Model Registry Architecture](./model_registry.md)
- [Worker Service Documentation](./worker_service.md)
- [Adding New Models Guide](./guides/ADDING_NEW_MODEL.md)
- [Model Package Refactor](./MODEL_PACKAGE_REFACTOR.md)

## Changelog

### Phase 1 (2025-10-08)
- ✅ Database migration for settings table
- ✅ Settings domain package with repository and service
- ✅ Admin HTTP endpoints for model management
- ✅ Admin UI page for model selection
- ✅ Unit tests for service layer
- ✅ Documentation

### Phase 2 (Planned)
- ⏳ Worker integration to read active model from DB
- ⏳ Role-based access control
- ⏳ Audit history logging
- ⏳ Per-project model overrides
- ⏳ Model cost tracking
- ⏳ Performance metrics dashboard
