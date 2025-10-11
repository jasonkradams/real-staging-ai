# Storage Reconciliation Runbook

This document describes how to run storage reconciliation to detect and fix mismatches between the database and S3/MinIO.

## Overview

The reconciliation system checks:
1. **Missing original files**: `original_url` in DB but file missing in S3
2. **Missing staged files**: `status=ready` with `staged_url` but file missing in S3

Images with missing files are marked as `status=error` with a descriptive error message.

## When to Run

- After S3/MinIO bucket recovery or migration
- When investigating orphaned DB records
- As part of periodic data integrity checks
- After suspected storage service outages

## Methods

### CLI Command (Recommended)

Run reconciliation from the command line using the built-in CLI:

```bash
# Dry-run (no changes applied)
make reconcile-images DRY_RUN=1

# Apply changes
make reconcile-images DRY_RUN=0

# With filters
make reconcile-images DRY_RUN=1 BATCH_SIZE=50 CONCURRENCY=10
```

**CLI Flags:**
- `--dry-run`: Don't apply changes (default: `false`)
- `--batch-size`: Number of images to check per batch (default: `100`)
- `--concurrency`: Number of concurrent S3 checks (default: `5`)
- `--project-id`: Optional UUID to filter by project
- `--status`: Optional status filter (`queued`, `processing`, `ready`, `error`)

**Example:**
```bash
# Check only ready images for a specific project
docker compose exec api go run ./cmd/reconcile/main.go \
  --project-id=b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12 \
  --status=ready \
  --dry-run=true \
  --batch-size=50
```

### Admin HTTP Endpoint (Alternative)

The reconciliation service is also exposed as an admin endpoint (requires `RECONCILE_ENABLED=1`):

```bash
# Enable the endpoint
export RECONCILE_ENABLED=1

# POST request with JWT auth
curl -X POST "http://localhost:8080/api/v1/admin/reconcile/images?dry_run=true&limit=100" \
  -H "Authorization: Bearer $(make token)" \
  -H "Content-Type: application/json"
```

**Query Parameters:**
- `project_id`: UUID (optional)
- `status`: string (optional)
- `limit`: integer, max 1000 (default: 100)
- `cursor`: UUID for pagination (optional)
- `dry_run`: boolean (default: false)
- `concurrency`: integer (default: 5)

**Response:**
```json
{
  "checked": 150,
  "missing_original": 2,
  "missing_staged": 1,
  "updated": 3,
  "dry_run": true,
  "examples": [
    {
      "image_id": "abc123...",
      "status": "ready",
      "error": "staged missing in storage"
    }
  ]
}
```

## Safety Mechanisms

1. **Dry-run mode**: Always test with `--dry-run=true` first
2. **Idempotent updates**: Safe to re-run; already-errored images are skipped
3. **Batch processing**: Configurable batch size to avoid long transactions
4. **Rate limiting**: Concurrency limit prevents S3 throttling
5. **Feature flag**: Admin endpoint requires `RECONCILE_ENABLED=1`

## Typical Workflow

1. **Investigate**: Run with `--dry-run=true` to see what would change
2. **Review**: Check the examples in output for expected mismatches
3. **Apply**: Run with `--dry-run=false` to update the database
4. **Verify**: Query affected images to confirm error messages are correct

```bash
# Step 1: Dry run
make reconcile-images DRY_RUN=1

# Step 2: Review output
# ... check examples and counts ...

# Step 3: Apply
make reconcile-images DRY_RUN=0

# Step 4: Verify in DB
docker compose exec postgres psql -U postgres -d realstaging \
  -c "SELECT id, status, error FROM images WHERE status='error' LIMIT 10;"
```

## Troubleshooting

### High missing_original count
- **Cause**: Bucket misconfiguration, wrong S3_ENDPOINT, or data loss
- **Action**: Verify S3 connection and bucket name; check S3 console

### High missing_staged count
- **Cause**: Worker failures or incomplete processing
- **Action**: Re-process images or investigate worker logs

### Timeout or slow performance
- **Cause**: High concurrency or large batch size
- **Action**: Reduce `--concurrency` and `--batch-size`

### "reconciliation is not enabled" error
- **Cause**: `RECONCILE_ENABLED` env var not set
- **Action**: Set `RECONCILE_ENABLED=1` in docker-compose.yml or environment

## Observability

- **Logs**: Structured JSON logs with `service=real-staging-api`
- **Traces**: OTEL span `reconcile.images` with attributes (`dry_run`, `limit`, etc.)
- **Metrics**: Log summaries include counts for checked/missing/updated

**Example log:**
```json
{
  "time": "2025-09-29T22:00:00Z",
  "level": "INFO",
  "msg": "reconcile: completed",
  "checked": 100,
  "missing_original": 2,
  "missing_staged": 1,
  "updated": 3,
  "dry_run": false
}
```

## Rollback

If reconciliation incorrectly marks images as errored:

1. Identify affected image IDs
2. Update status back to previous value:
   ```sql
   UPDATE images
   SET status = 'ready', error = NULL, updated_at = now()
   WHERE id IN ('uuid1', 'uuid2', ...);
   ```
3. Investigate root cause (S3 endpoint misconfiguration, transient errors, etc.)
4. Re-run reconciliation after fixing the issue

## Automation

For periodic checks, consider:
- **Nightly cron**: Run dry-run and alert on high missing counts
- **Post-deployment**: Run after infrastructure changes
- **Monitoring integration**: Export metrics to Prometheus/Datadog

**Example cron:**
```bash
# Daily at 2 AM, dry-run and email results
0 2 * * * cd /app && make reconcile-images DRY_RUN=1 | mail -s "Reconciliation Report" ops@example.com
```
