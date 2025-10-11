# Auth0 Integration Setup

## Changes Made

### ✅ Removed Legacy Token System
- **Deleted**: `components/TokenBar.tsx` (manual token entry)
- **Updated**: `components/AuthButton.tsx` to use `/auth/login` and `/auth/logout`

### ✅ Auth0 SDK Endpoints (Automatic)
- All auth routes are automatically provided by Auth0 SDK v4 middleware:
  - `/auth/login` - Redirects to Auth0 Universal Login
  - `/auth/logout` - Clears session and logs out
  - `/auth/callback` - Handles OAuth callback
  - `/auth/access-token` - Provides JWT for API calls (requires `enableAccessTokenEndpoint: true`)
- No manual route creation needed!

### ✅ Existing Auth0 Integration (Already Working)
- `lib/api.ts` - `apiFetch()` function fetches token from `/auth/access-token`
- `components/SSEViewer.tsx` - Fetches token for EventSource connections
- `middleware.ts` - Protects routes and handles Auth0 authentication
- All pages use `apiFetch()` for API calls with automatic token injection

## How It Works

### Client-Side Flow
1. User clicks "Login" → Redirects to `/auth/login`
2. Auth0 middleware handles login flow
3. User redirected back to `/auth/callback`
4. Session cookie created
5. User can access protected routes

### API Call Flow
1. Page calls `apiFetch('/v1/projects')`
2. `apiFetch` calls `/auth/access-token` to get JWT
3. JWT added to `Authorization: Bearer <token>` header
4. Request sent to backend API
5. Backend validates JWT via Auth0

### SSE Flow
1. Component calls `/auth/access-token` to get JWT
2. Token added as query parameter: `/api/v1/events?image_id=xxx&access_token=<token>`
3. Backend validates token from query parameter (EventSource can't set headers)

## Required Environment Variables

Ensure your `.env.local` has:

```bash
# Auth0 Configuration
AUTH0_DOMAIN=dev-sleeping-pandas.us.auth0.com
AUTH0_CLIENT_ID=your_client_id
AUTH0_CLIENT_SECRET=your_client_secret
AUTH0_SECRET=your_random_secret_32_chars
AUTH0_AUDIENCE=https://api.realstaging.local
APP_BASE_URL=http://localhost:3000

# API
NEXT_PUBLIC_API_BASE=/api

# Optional: AUTH0_SCOPE defaults to "openid profile email offline_access"
```

## Testing

1. Start the app: `make up`
2. Visit http://localhost:3000
3. Click "Sign In" 
4. Login via Auth0 Universal Login
5. Try creating a project - should work now!

## Troubleshooting

If you get 401 errors:
- Check `/auth/access-token` returns a token
- Check browser DevTools → Network → Headers for API requests
- Verify `Authorization: Bearer <token>` header is present
- Check backend logs to see if token validation fails
