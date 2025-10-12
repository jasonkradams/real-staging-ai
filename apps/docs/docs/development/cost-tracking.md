# Cost Tracking Feature

**Status**: ✅ Backend Implemented  
**Date**: 2025-10-08

## Overview

The cost tracking feature tracks the cost of processing each image and provides aggregated cost summaries at the project level. This enables users to monitor spending and helps with billing and analytics.

## Architecture

### Database Schema

**New columns in `images` table**:

```sql
ALTER TABLE images
  ADD COLUMN cost_usd DECIMAL(10, 4) DEFAULT 0.00,
  ADD COLUMN model_used VARCHAR(255),
  ADD COLUMN processing_time_ms INTEGER,
  ADD COLUMN replicate_prediction_id VARCHAR(255);
```

**Indexes**:

- `idx_images_project_cost` - For fast cost aggregation by project
- `idx_images_model` - For cost analysis by model

### Cost Calculation

**Model Pricing Table** (`apps/worker/internal/staging/pricing.go`):

```go
var modelPricingTable = []ModelPricing{
    {
        ModelID:      model.ModelQwenImageEdit,
        CostPerImage: 0.005, // $0.005 per image
    },
    {
        ModelID:      model.ModelFluxKontextMax,
        CostPerImage: 0.03, // $0.03 per image
    },
}
```

**Note**: Prices are estimates based on Replicate's pricing. Actual costs may vary based on processing time and model usage.

### API Endpoints

#### Get Project Cost Summary

**Endpoint**: `GET /api/v1/projects/:project_id/cost`

**Response**:

```json
{
  "project_id": "123e4567-e89b-12d3-a456-426614174000",
  "total_cost_usd": 0.15,
  "image_count": 5,
  "avg_cost_usd": 0.03
}
```

**Fields**:

- `total_cost_usd` - Total cost for all images in the project
- `image_count` - Number of images processed
- `avg_cost_usd` - Average cost per image

#### Individual Image Costs

Image cost information is included in the image response:

```json
{
  "id": "...",
  "project_id": "...",
  "status": "ready",
  "cost_usd": 0.03,
  "model_used": "black-forest-labs/flux-kontext-max",
  "processing_time_ms": 12500,
  "replicate_prediction_id": "abc123..."
}
```

## Implementation Flow

### 1. Image Processing (Worker)

When the worker processes an image:

```go
// 1. Get model pricing
cost := staging.GetModelCost(modelID)

// 2. Process image with Replicate
prediction, err := replicateClient.Run(...)

// 3. Update cost in database
err = imageRepo.UpdateImageCost(ctx, imageID, cost, string(modelID),
    processingTimeMs, prediction.ID)
```

### 2. Cost Retrieval (API)

Users can retrieve cost summaries:

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/projects/{project_id}/cost
```

## Data Model

### Image Model

```go
type Image struct {
    // ... existing fields
    CostUSD               *float64  `json:"cost_usd,omitempty"`
    ModelUsed             *string   `json:"model_used,omitempty"`
    ProcessingTimeMs      *int      `json:"processing_time_ms,omitempty"`
    ReplicatePredictionID *string   `json:"replicate_prediction_id,omitempty"`
}
```

### Project Cost Summary

```go
type ProjectCostSummary struct {
    ProjectID    uuid.UUID `json:"project_id"`
    TotalCostUSD float64   `json:"total_cost_usd"`
    ImageCount   int       `json:"image_count"`
    AvgCostUSD   float64   `json:"avg_cost_usd"`
}
```

## Usage Examples

### Get Project Cost via API

```bash
# Get auth token
TOKEN=$(make token)

# Get project cost summary
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/projects/123e4567-e89b-12d3-a456-426614174000/cost
```

### Query Costs Directly

```sql
-- Get total cost for a project
SELECT
    project_id,
    SUM(cost_usd) as total_cost,
    COUNT(*) as image_count,
    AVG(cost_usd) as avg_cost
FROM images
WHERE project_id = '123e4567-e89b-12d3-a456-426614174000'
GROUP BY project_id;

-- Get cost breakdown by model
SELECT
    model_used,
    COUNT(*) as image_count,
    SUM(cost_usd) as total_cost,
    AVG(cost_usd) as avg_cost
FROM images
GROUP BY model_used;
```

## Migration

### Apply Migration

**Development**:

```bash
make migrate
```

**Test**:

```bash
make migrate-test
```

### Rollback

```bash
# Development
make migrate-down-dev

# Or specific migration
migrate down 1
```

## Future Enhancements

### Phase 2: Advanced Cost Features

- [ ] Monthly cost reports
- [ ] Cost alerts and budgets
- [ ] Detailed billing statements
- [ ] Cost breakdown by user
- [ ] Export cost data to CSV/PDF
- [ ] Integration with Stripe for billing

### Phase 3: Cost Optimization

- [ ] Real-time cost estimation before processing
- [ ] Model recommendation based on cost vs quality
- [ ] Bulk processing discounts
- [ ] Cost prediction based on image complexity

## Pricing Updates

To update model pricing:

1. Edit `apps/worker/internal/staging/pricing.go`
2. Update the `modelPricingTable` array
3. Redeploy worker service

```go
var modelPricingTable = []ModelPricing{
    {
        ModelID:      model.ModelQwenImageEdit,
        CostPerImage: 0.006, // Updated price
    },
    // ...
}
```

## Analytics Queries

### Cost by Time Period

```sql
SELECT
    DATE_TRUNC('day', created_at) as day,
    SUM(cost_usd) as daily_cost,
    COUNT(*) as images_processed
FROM images
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY day
ORDER BY day DESC;
```

### Top Spending Projects

```sql
SELECT
    p.id,
    p.name,
    SUM(i.cost_usd) as total_cost,
    COUNT(i.id) as image_count
FROM projects p
JOIN images i ON i.project_id = p.id
GROUP BY p.id, p.name
ORDER BY total_cost DESC
LIMIT 10;
```

### Model Usage and Cost

```sql
SELECT
    model_used,
    COUNT(*) as usage_count,
    SUM(cost_usd) as total_cost,
    AVG(cost_usd) as avg_cost,
    AVG(processing_time_ms) as avg_processing_time
FROM images
WHERE model_used IS NOT NULL
GROUP BY model_used;
```

## Testing

### Unit Tests

```bash
cd apps/api
go test ./internal/image -v -run TestGetProjectCostSummary
```

### Integration Tests

```bash
# Start dev environment
make up

# Test cost endpoint
TOKEN=$(make token)
PROJECT_ID="..." # Use actual project ID

curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/projects/${PROJECT_ID}/cost"
```

## Monitoring

### Key Metrics to Track

- Average cost per image
- Total daily/monthly spending
- Cost per model
- Failed predictions (no cost charged)
- Processing time vs cost correlation

### Alerts

Set up alerts for:

- Daily spending exceeds threshold
- Unusual cost spikes
- Projects with abnormally high costs
- Failed predictions with cost > 0

## Security Considerations

- Cost data is sensitive - restrict access via JWT
- Only project owners should see project costs
- Admin users can see all costs
- Audit log for cost queries (future)

## Related Documentation

- [Model Registry](./model_registry.md)
- [Model Package Refactor](./MODEL_PACKAGE_REFACTOR.md)
- [Admin Model Selection](./ADMIN_MODEL_SELECTION.md)
- [Worker Service](./worker_service.md)

## Changelog

### 2025-10-08

- ✅ Database migration for cost tracking columns
- ✅ Image model updated with cost fields
- ✅ Repository methods for cost updates and aggregation
- ✅ API endpoint for project cost summary
- ✅ Model pricing configuration
- ✅ Documentation

### Future

- ⏳ UI components to display costs
- ⏳ Cost analytics dashboard
- ⏳ Billing integration
- ⏳ Cost alerts and budgets
