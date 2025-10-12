# Your First Project

This tutorial walks you through creating your first virtual staging project, from uploading an image to receiving the staged result.

## Prerequisites

- [Installation](installation.md) completed
- Development stack running (`make up`)
- Auth0 token (generate with `make token`)

## Step 1: Authenticate

Get an access token for API requests:

```bash
export TOKEN=$(make token)
```

Or use the web UI at [http://localhost:3000](http://localhost:3000) to log in via Auth0.

## Step 2: Create a Project

Projects organize your images. Create one via API:

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My First Staging Project",
    "description": "Learning Real Staging AI"
  }' | jq
```

**Response:**
```json
{
  "id": "01J9XYZ123ABC456DEF789GH",
  "name": "My First Staging Project",
  "description": "Learning Real Staging AI",
  "created_at": "2025-10-12T20:06:00Z",
  "updated_at": "2025-10-12T20:06:00Z"
}
```

Save the project ID:
```bash
export PROJECT_ID="01J9XYZ123ABC456DEF789GH"
```

## Step 3: Request Presigned Upload URL

Get a secure presigned URL to upload your image to S3:

```bash
curl -X POST http://localhost:8080/api/v1/uploads/presign \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "'$PROJECT_ID'",
    "filename": "empty-room.jpg",
    "content_type": "image/jpeg"
  }' | jq
```

**Response:**
```json
{
  "upload_url": "http://localhost:9000/real-staging/uploads/...",
  "key": "uploads/user_123/empty-room-uuid.jpg",
  "expires_at": "2025-10-12T21:06:00Z"
}
```

Save the key:
```bash
export S3_KEY="uploads/user_123/empty-room-uuid.jpg"
```

## Step 4: Upload Your Image

Use the presigned URL to upload an image directly to S3:

```bash
# Example: upload a local image
curl -X PUT "http://localhost:9000/real-staging/uploads/..." \
  -H "Content-Type: image/jpeg" \
  --data-binary "@/path/to/your/empty-room.jpg"
```

!!! tip "Web Interface"
    The web UI at http://localhost:3000 handles this automatically with drag-and-drop upload.

## Step 5: Create Image Job

Request AI staging for your uploaded image:

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "'$PROJECT_ID'",
    "original_url": "s3://real-staging/'$S3_KEY'",
    "room_type": "living_room",
    "style": "modern"
  }' | jq
```

**Response:**
```json
{
  "id": "01J9XYZ789ABC123DEF456GH",
  "project_id": "01J9XYZ123ABC456DEF789GH",
  "status": "queued",
  "room_type": "living_room",
  "style": "modern",
  "created_at": "2025-10-12T20:08:00Z"
}
```

Save the image ID:
```bash
export IMAGE_ID="01J9XYZ789ABC123DEF456GH"
```

## Step 6: Monitor Processing

The worker processes your job asynchronously. Check status:

```bash
curl http://localhost:8080/api/v1/images/$IMAGE_ID \
  -H "Authorization: Bearer $TOKEN" | jq
```

Status transitions:
1. `queued` - Job is in queue
2. `processing` - Worker is staging the image
3. `ready` - Staged image is available
4. `error` - Something went wrong

## Step 7: Get Results

Once status is `ready`, download the staged image:

```bash
curl http://localhost:8080/api/v1/images/$IMAGE_ID \
  -H "Authorization: Bearer $TOKEN" | jq
```

**Response:**
```json
{
  "id": "01J9XYZ789ABC123DEF456GH",
  "status": "ready",
  "original_url": "s3://real-staging/uploads/...",
  "staged_url": "s3://real-staging/staged/...",
  "room_type": "living_room",
  "style": "modern",
  "processing_time_ms": 8945,
  "cost_cents": 1
}
```

Get a presigned download URL:

```bash
curl http://localhost:8080/api/v1/images/$IMAGE_ID/download \
  -H "Authorization: Bearer $TOKEN" | jq .url
```

## Real-time Updates with SSE

Instead of polling, use Server-Sent Events for real-time updates:

```bash
curl -N http://localhost:8080/api/v1/events \
  -H "Authorization: Bearer $TOKEN"
```

You'll receive events as they happen:
```
data: {"type":"image.processing","image_id":"01J9XYZ789ABC123DEF456GH"}

data: {"type":"image.ready","image_id":"01J9XYZ789ABC123DEF456GH"}
```

## Batch Upload (Advanced)

Upload multiple images at once:

```bash
curl -X POST http://localhost:8080/api/v1/images/batch \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "'$PROJECT_ID'",
    "images": [
      {
        "original_url": "s3://real-staging/'$S3_KEY1'",
        "room_type": "living_room",
        "style": "modern"
      },
      {
        "original_url": "s3://real-staging/'$S3_KEY2'",
        "room_type": "bedroom",
        "style": "contemporary"
      }
    ],
    "default_room_type": "living_room",
    "default_style": "modern"
  }' | jq
```

All images are queued and processed concurrently.

## Web Interface Workflow

Using the web UI at http://localhost:3000:

1. **Log In** - Click "Sign In" and authenticate via Auth0
2. **Upload** - Go to `/upload` and drag-drop images
3. **Configure** - Set room type and style
4. **Submit** - Images are uploaded and jobs created automatically
5. **Monitor** - View real-time progress on `/images` page
6. **Download** - Click download button when ready

## What You've Learned

✅ Create projects to organize your work  
✅ Request presigned URLs for secure uploads  
✅ Upload images directly to S3  
✅ Create image jobs for AI staging  
✅ Monitor job status  
✅ Download staged results  
✅ Use SSE for real-time updates  

## Next Steps

- [Learn about Configuration](../guides/configuration.md)
- [Explore the Architecture](../architecture/)
- [Read the API Reference](../api-reference/)
- [Set up Authentication](../guides/authentication.md)

## Troubleshooting

### Job stays in "queued" status

**Cause:** Worker might not be running or can't connect to Redis.

**Solution:**
```bash
# Check worker logs
docker compose logs worker

# Restart worker
docker compose restart worker
```

### Job moves to "error" status

**Cause:** Invalid Replicate token, network issues, or invalid image.

**Solution:**
```bash
# Check worker logs for error details
docker compose logs worker | grep ERROR

# Verify Replicate token
echo $REPLICATE_API_TOKEN
```

### Upload fails with 403 Forbidden

**Cause:** Presigned URL expired or invalid.

**Solution:**
- Presigned URLs expire after 15 minutes
- Request a new presigned URL and try again

### Can't download staged image

**Cause:** Image might still be processing or job failed.

**Solution:**
```bash
# Check image status
curl http://localhost:8080/api/v1/images/$IMAGE_ID \
  -H "Authorization: Bearer $TOKEN" | jq .status
```

---

Next: [Configuration Guide →](../guides/configuration.md)
