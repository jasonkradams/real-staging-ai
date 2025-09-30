# Phase 1 Frontend Implementation

## Overview

The Phase 1 frontend is a minimal Next.js application that validates the core backend flows without requiring a full OAuth implementation. It provides a functional UI for image upload, job tracking, and real-time status updates.

## Tech Stack

- **Framework**: Next.js 15.5.4 (App Router)
- **Language**: TypeScript 5.6.2
- **Styling**: Tailwind CSS 3.4.13
- **Node**: 18.18.0+

## Features Implemented

### 1. Pages

#### Dashboard (`/`)
- Landing page with navigation cards
- Links to Upload and Images pages

#### Upload Page (`/upload`)
- Project management:
  - Create new projects
  - Select from existing projects
  - Refresh project list
- Image upload flow:
  1. Request presigned S3 URL from API (`POST /v1/uploads/presign`)
  2. Upload file directly to S3 via presigned URL
  3. Create image record in database (`POST /v1/images`)
- Real-time status feedback during upload

#### Images Page (`/images`)
- Project selector with filtering
- Image list table with:
  - Image ID (clickable to select)
  - Status badges (queued, processing, ready, error)
  - Links to view original/staged images (via presigned URLs)
  - Created/updated timestamps
- Selected image detail panel
- Live SSE viewer for real-time job updates

### 2. Components

#### TokenBar
- Simple token management UI in header
- Stores bearer token in `localStorage`
- Used for all authenticated API requests
- Provides "Set Token" / "Edit Token" interface

#### SSEViewer
- Client-side EventSource implementation
- Connects to `/api/v1/events?image_id=<id>&access_token=<token>`
- Listens for `job_update` events
- Displays event log with timestamps
- Auto-refreshes image list when job completes (ready/error status)

### 3. API Client (`lib/api.ts`)

```typescript
export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T>
```

- Centralized API client with:
  - Automatic bearer token injection from `localStorage`
  - JSON request/response handling
  - Error handling with status codes
  - TypeScript generics for type safety

### 4. Authentication Approach

**Phase 1: Manual Token Entry**
- Users paste a bearer token obtained via `make token`
- Token stored in browser `localStorage`
- Added to requests via `Authorization: Bearer <token>` header
- SSE uses query parameter (`access_token=<token>`) since EventSource doesn't support custom headers

**Rationale:**
- Validates backend auth flows without OAuth complexity
- Allows rapid iteration on core features
- Sufficient for Phase 1 testing and development

**Phase 2 Plan:**
- Integrate `@auth0/nextjs-auth0` SDK (Note: v4 has breaking API changes from v3)
- Implement proper OAuth login/logout flow
- Add protected routes and session management
- Auto-refresh expired tokens

## API Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/projects` | GET | List all projects |
| `/v1/projects` | POST | Create new project |
| `/v1/projects/:id/images` | GET | List images for project |
| `/v1/images` | POST | Create image record |
| `/v1/images/:id/presign` | GET | Get presigned URL for viewing |
| `/v1/uploads/presign` | POST | Get presigned URL for upload |
| `/v1/events` | GET (SSE) | Stream job status updates |

## Environment Variables

```bash
# .env.local
NEXT_PUBLIC_API_BASE=/api  # API base URL (defaults to /api)
```

The Next.js app proxies `/api/*` to `http://localhost:8080/api/*` via `next.config.js` rewrites.

## Development Workflow

```bash
# Start API and worker
make up

# In separate terminal: Start Next.js dev server
cd apps/web
npm run dev
# Opens http://localhost:3000

# Generate auth token
make token
# Paste token into "Set Token" field in UI
```

## Build & Deploy

```bash
cd apps/web
npm run build  # Production build
npm start      # Start production server
```

## Testing Checklist

- [x] Create project
- [x] Select project from dropdown
- [x] Upload image file
- [x] View original image (presigned URL)
- [x] Connect SSE viewer to image
- [x] Receive `job_update` events in real-time
- [x] View staged image when ready
- [x] Handle error states
- [x] Refresh lists work correctly

## Known Limitations

1. **No OAuth flow**: Users must manually obtain and paste tokens
2. **No token refresh**: Expired tokens require manual replacement
3. **No protected routes**: All pages publicly accessible (auth happens at API level)
4. **No user profile**: No UI for user information or account management
5. **Limited error handling**: Some edge cases may not have user-friendly messages
6. **No loading skeletons**: Uses simple "Loading..." text states

## Next Steps (Phase 2)

1. Integrate Auth0 SDK for proper OAuth
2. Add protected route middleware
3. Implement token refresh logic
4. Add user profile dropdown
5. Improve loading states with skeleton screens
6. Add form validation with better error messages
7. Add image preview before upload
8. Add bulk operations (delete, re-process)
9. Add pagination for large image lists
10. Add search/filter functionality

## File Structure

```
apps/web/
├── app/
│   ├── layout.tsx          # Root layout with header
│   ├── page.tsx            # Dashboard
│   ├── upload/
│   │   └── page.tsx        # Upload page
│   ├── images/
│   │   └── page.tsx        # Images page
│   └── globals.css         # Tailwind styles
├── components/
│   ├── TokenBar.tsx        # Token management UI
│   └── SSEViewer.tsx       # SSE client component
├── lib/
│   └── api.ts              # API client library
├── next.config.js          # Next.js config with API proxy
├── tailwind.config.ts      # Tailwind config
└── package.json            # Dependencies
```

## Dependencies

```json
{
  "dependencies": {
    "next": "^15.5.4",
    "react": "18.2.0",
    "react-dom": "18.2.0"
  },
  "devDependencies": {
    "@types/node": "20.14.12",
    "@types/react": "18.2.41",
    "@types/react-dom": "18.2.18",
    "autoprefixer": "10.4.20",
    "eslint": "8.57.0",
    "eslint-config-next": "^15.5.4",
    "postcss": "8.4.49",
    "tailwindcss": "3.4.13",
    "typescript": "5.6.2"
  }
}
```
