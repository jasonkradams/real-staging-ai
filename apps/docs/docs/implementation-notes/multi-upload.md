# Multi-File Upload Feature

## Overview

The multi-file upload feature allows users to upload and stage multiple property images simultaneously with intelligent defaults and per-image overrides. This significantly improves workflow efficiency for users staging entire properties.

## User Experience

### Key Features

1. **Drag & Drop Multiple Files**: Users can drag and drop up to 50 images at once
2. **Default Settings**: Set room type and furniture style that apply to all images
3. **Individual Overrides**: Expand any image to set custom room type or style
4. **Real-time Progress**: See upload progress for each image individually
5. **Partial Success Handling**: Continue processing even if some images fail
6. **Visual Feedback**: Color-coded status indicators for each file

### UI Components

#### File Selection Zone
- Accepts multiple files via drag & drop or click to browse
- Visual feedback during drag operation
- Displays count of selected files
- Supports JPG, PNG, and WEBP formats
- Max 10MB per file

#### File List
- Scrollable list showing all selected files
- File name, size, and status for each
- Remove button to deselect individual files
- Expand/collapse for per-file settings

#### Default Settings
- Room Type dropdown (applies to all images without overrides)
- Furniture Style dropdown (applies to all images without overrides)
- Clearly labeled as "default" to distinguish from overrides

#### Progress Indicators
- Per-file progress bars showing: presigning → uploading → creating
- Success/error indicators
- Detailed error messages for failed uploads
- Success summary showing count of successfully queued images

## API Endpoints

### Batch Create Images

**Endpoint**: `POST /api/v1/images/batch`

**Description**: Create multiple staging images in a single request. Each image is processed independently, allowing for partial success.

**Request Body**:
```json
{
  "images": [
    {
      "project_id": "uuid",
      "original_url": "https://...",
      "room_type": "bedroom",      // optional, per-image override
      "style": "modern"             // optional, per-image override
    },
    {
      "project_id": "uuid",
      "original_url": "https://..."
      // Uses defaults if room_type/style not specified
    }
  ]
}
```

**Validation**:
- Minimum: 1 image
- Maximum: 50 images per batch
- Each image validated individually
- Returns 422 with detailed field-level errors

**Response Codes**:
- `201 Created`: All images created successfully
- `207 Multi-Status`: Partial success (some succeeded, some failed)
- `400 Bad Request`: All images failed
- `422 Unprocessable Entity`: Validation errors

**Success Response**:
```json
{
  "images": [
    {
      "id": "uuid",
      "project_id": "uuid",
      "original_url": "https://...",
      "status": "queued",
      // ... other image fields
    }
  ],
  "errors": [
    {
      "index": 2,
      "message": "failed to create image: ..."
    }
  ],
  "success": 9,
  "failed": 1
}
```

## Frontend Implementation

### State Management

The upload page maintains:
- `files`: Array of `FileWithOverrides` containing File object, metadata, and override settings
- `defaultRoomType`: Default room type for all files
- `defaultStyle`: Default style for all files
- `uploadProgress`: Map of file ID to upload progress/status

### Upload Flow

1. **File Selection**: User selects multiple files
2. **Configure Defaults**: User sets default room type/style (optional)
3. **Configure Overrides**: User expands individual files to set custom settings (optional)
4. **Submit**: User clicks "Upload & Stage"
5. **Concurrent Upload**: All files upload concurrently (Promise.all)
   - For each file:
     - Presign upload URL
     - Upload to S3
     - Create image record
6. **Progress Display**: Real-time status updates for each file
7. **Results**: Summary showing success/failure counts with links to image gallery

### Override Logic

For each image:
```typescript
const roomType = fileData.roomType || defaultRoomType
const style = fileData.style || defaultStyle
```

Individual file settings take precedence over defaults.

## Backend Implementation

### Service Layer

**BatchCreateImages Method**:
- Iterates through each CreateImageRequest
- Calls CreateImage for each (reuses existing logic)
- Collects successful results and errors
- Returns aggregated response
- Logs batch completion statistics

**Design Rationale**:
- Reuses existing CreateImage validation and business logic
- Each image creates its own job and enqueues independently
- Failures are isolated - one image failure doesn't block others
- Simple, maintainable implementation

### Handler Layer

**BatchCreateImages Handler**:
- Validates batch size (1-50 images)
- Validates each image individually
- Returns field-specific errors (e.g., `images[2].project_id`)
- Maps response to appropriate HTTP status code
- Returns 207 Multi-Status for partial success

### Database Impact

- No schema changes required
- Uses existing `images` and `jobs` tables
- Each image is a separate transaction
- Failures don't rollback successful images

## Performance Considerations

### Concurrency

**Frontend**:
- Uploads happen concurrently (Promise.all)
- Limited by browser connection limits (~6-10 concurrent)
- S3 presigned URLs avoid API bottleneck

**Backend**:
- Each image processed sequentially in batch request
- Each image creates independent job for worker
- Worker processes jobs concurrently (configured concurrency)

### Limits

- **Batch Size**: 50 images per request
  - Balances UX convenience with request size
  - Prevents timeout issues
  - Users can submit multiple batches if needed

- **File Size**: 10MB per file (existing limit)

- **Total Request Size**: ~500MB theoretical max (50 × 10MB)
  - Actual JSON payload is small (<50KB for 50 images)
  - Large files uploaded directly to S3

### Recommendations

For properties with 50+ images:
1. Upload in batches of 25-50
2. Monitor progress of each batch
3. All images appear in same project regardless of batch

## Testing

### Frontend Tests (`apps/web/app/upload/__tests__/multi-upload.test.tsx`)

- ✅ Renders multi-file upload interface
- ✅ Supports selecting multiple files
- ✅ Allows removing individual files
- ✅ Shows default settings section
- ✅ Allows expanding files for custom overrides
- ✅ Validates project must be selected
- ✅ Uploads multiple files with defaults
- ✅ Handles upload errors gracefully
- ✅ Applies individual overrides correctly

### Backend Tests

**Handler Tests** (`apps/api/internal/image/default_handler_batch_test.go`):
- ✅ Success: All images created
- ✅ Partial success: Some images fail
- ✅ Empty request validation
- ✅ Too many images validation (>50)
- ✅ Invalid image data validation
- ✅ Proper HTTP status codes (201, 207, 400, 422)

**Service Tests** (`apps/api/internal/image/default_service_batch_test.go`):
- ✅ All images created successfully
- ✅ Partial failure handling
- ✅ Empty request handling
- ✅ Correct aggregation of results

## Migration Guide

### For Existing Users

No migration needed - single file upload still works exactly as before. Multi-upload is an enhancement, not a replacement.

### API Compatibility

- **New endpoint**: `POST /api/v1/images/batch`
- **Existing endpoint unchanged**: `POST /api/v1/images`
- Clients can use either endpoint based on need
- Single image uploads should continue using `/images` for simplicity

## Future Enhancements

Potential improvements for future iterations:

1. **Resume Failed Uploads**: Allow re-uploading only failed images from a batch
2. **Bulk Edit**: Apply room type/style changes to multiple selected files at once
3. **Template Presets**: Save common room type/style combinations
4. **Increased Batch Size**: Support 100+ images with streaming/pagination
5. **ZIP Upload**: Upload a ZIP file that auto-extracts to multiple images
6. **Folder Upload**: Preserve folder structure when uploading multiple files
7. **Progress Persistence**: Save upload progress across page refreshes

## Troubleshooting

### Common Issues

**Q: Why did only some of my images upload?**

A: The batch upload continues processing even if individual images fail. Check the error details for each failed image. Common causes:
- File too large (>10MB)
- Invalid project selected
- Network timeout during upload
- S3 connectivity issues

**Q: Can I upload more than 50 images?**

A: Currently limited to 50 images per batch. For larger sets, submit multiple batches. All images will appear in the same project.

**Q: Do overrides work for all images?**

A: Overrides only apply to the specific image you configure. Use default settings to apply room type/style to all images without individual overrides.

**Q: What happens if I close the page during upload?**

A: Uploads that completed before closing will be processed. In-progress uploads will fail and need to be re-uploaded.

## Related Documentation

- [API Documentation](../apps/api/web/api/v1/oas3.yaml) - OpenAPI specification
- [Upload Flow](./frontend/PHASE1_IMPLEMENTATION.md) - Original upload implementation
- [Testing Guide](./guides/TDD_GUIDE.md) - Testing best practices
