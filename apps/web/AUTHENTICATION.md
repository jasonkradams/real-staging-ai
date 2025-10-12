# Authentication & Route Protection

This document describes the authentication and authorization implementation for the web application.

## Protected Routes

### User Routes (Redirect to Login)
- `/upload` - Image upload page
- `/images` - Image gallery page

**Behavior for unauthenticated users:**
- Navigation links are hidden in the UI
- Direct navigation redirects to `/auth/login` with `returnTo` parameter
- After login, users are redirected back to their intended destination

### Admin Routes (404 Response)
- `/admin/*` - All admin pages including `/admin/settings`

**Behavior for unauthenticated users:**
- Returns HTTP 404 (Not Found) to hide existence from unauthorized users
- No navigation links shown in UI
- No redirect to login (security through obscurity)

## Implementation

### Client-Side Protection
- **ProtectedNav Component** (`components/ProtectedNav.tsx`)
  - Uses Auth0's `useUser()` hook to check authentication
  - Conditionally renders navigation links only for authenticated users
  - Returns `null` when user is not authenticated

### Server-Side Protection
- **Middleware** (`middleware.ts`)
  - Checks authentication via Auth0 session before allowing access
  - Two-tier protection strategy:
    1. **User routes**: Redirect to login for better UX
    2. **Admin routes**: Return 404 to hide existence
  - Prevents direct URL access even if users bypass client-side checks

### Layout Updates
- **Root Layout** (`app/layout.tsx`)
  - Uses `ProtectedNav` component for conditional navigation
  - Removed hardcoded protected links from footer
  - Only shows protected navigation to authenticated users

## Testing Authentication

### As Unauthenticated User
```bash
# Navigate to user routes - should redirect to login
curl -I http://localhost:3000/upload
# Response: 302 Redirect to /auth/login?returnTo=/upload

# Navigate to admin routes - should return 404
curl -I http://localhost:3000/admin/settings
# Response: 404 Not Found
```

### As Authenticated User
1. Login via `/auth/login`
2. Navigation links to Upload and Images appear in header
3. Can access `/upload`, `/images`, and `/admin/*` routes
4. Session is maintained via Auth0 cookies

## Security Considerations

1. **Defense in Depth**: Both client-side (hidden links) and server-side (middleware) protection
2. **Admin Obscurity**: 404 response for admin routes prevents reconnaissance
3. **Graceful UX**: User routes redirect to login with return path
4. **Session Management**: Handled by Auth0 SDK with rolling sessions
5. **Token Injection**: API calls automatically include Auth0 access token via `apiFetch()`
