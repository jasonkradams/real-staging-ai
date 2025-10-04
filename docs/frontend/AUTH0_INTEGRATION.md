# Auth0 SDK Integration

## Overview

The Virtual Staging AI frontend now uses the official `@auth0/nextjs-auth0` SDK (v4) to handle user authentication. This replaces the manual token entry from Phase 1 with a proper OAuth 2.0 / OpenID Connect flow.

## Features

- **Universal Login**: Users authenticate via Auth0's hosted login page
- **Automatic Token Management**: Access tokens are stored in secure, httpOnly session cookies
- **Token Refresh**: Access tokens are automatically refreshed before expiration
- **Protected Routes**: Middleware automatically redirects unauthenticated users to login
- **SSO Support**: Single sign-on across Auth0 applications
- **Secure by Default**: CSRF protection, httpOnly cookies, and encrypted sessions

## Architecture

### Components

```
apps/web/
├── lib/
│   ├── auth0.ts              # Auth0 SDK client instance
│   └── api.ts                # API client with automatic token injection
├── middleware.ts             # Auth0 middleware for session management
├── components/
│   ├── UserProvider.tsx      # Wraps app with Auth0Provider
│   ├── AuthButton.tsx        # Login/logout UI component
│   └── SSEViewer.tsx         # SSE client (uses access_token query param)
└── app/
    └── layout.tsx            # Root layout with UserProvider
```

### Authentication Flow

1. **Login Initiation**: User clicks "Login" → redirected to `/auth/login`
2. **Auth0 Universal Login**: User authenticates on Auth0's hosted page
3. **Callback**: Auth0 redirects to `/auth/callback` with authorization code
4. **Session Creation**: Middleware exchanges code for tokens, creates encrypted session cookie
5. **Access Protected Routes**: Middleware validates session on each request
6. **API Requests**: Frontend fetches access token from `/auth/access-token` endpoint
7. **Token Refresh**: SDK automatically refreshes expired tokens using refresh token

### Session Management

- **Storage**: Encrypted, httpOnly session cookies (not accessible to JavaScript)
- **Duration**: Rolling sessions (7 days by default)
- **Refresh**: Tokens refreshed automatically when expired
- **Security**: CSRF protection, secure flag in production, SameSite=Lax

## Configuration

### Environment Variables

Create a `.env.local` file in `apps/web/`:

```bash
# Copy from env.example
cp env.example .env.local
```

Required variables:

| Variable               | Description                                       | Example                              |
| ---------------------- | ------------------------------------------------- | ------------------------------------ |
| `AUTH0_DOMAIN`         | Your Auth0 tenant domain                          | `dev-sleeping-pandas.us.auth0.com`   |
| `AUTH0_CLIENT_ID`      | Auth0 application client ID                       | `abc123...`                          |
| `AUTH0_CLIENT_SECRET`  | Auth0 application client secret                   | `xyz789...`                          |
| `AUTH0_SECRET`         | Secret for encrypting session cookies (≥32 chars) | Generate with `openssl rand -hex 32` |
| `APP_BASE_URL`         | Application base URL                              | `http://localhost:3000`              |
| `AUTH0_AUDIENCE`       | Auth0 API audience (must match backend)           | `https://api.virtualstaging.local`   |
| `AUTH0_SCOPE`          | OAuth scopes                                      | `openid profile email`               |
| `NEXT_PUBLIC_API_BASE` | Backend API base URL                              | `/api`                               |

### Auth0 Application Settings

In the Auth0 Dashboard, configure your application:

1. **Application Type**: Regular Web Application
2. **Allowed Callback URLs**:
   - Development: `http://localhost:3000/auth/callback`
   - Production: `https://yourdomain.com/auth/callback`
3. **Allowed Logout URLs**:
   - Development: `http://localhost:3000`
   - Production: `https://yourdomain.com`
4. **Allowed Web Origins**:
   - Development: `http://localhost:3000`
   - Production: `https://yourdomain.com`

### Backend API Configuration

The backend API must be configured with the same Auth0 settings:

- `AUTH0_DOMAIN`: Same as frontend
- `AUTH0_AUDIENCE`: Same as frontend (e.g., `https://api.virtualstaging.local`)

The API validates JWT access tokens signed by Auth0.

## Usage

### Protecting Pages

All pages under `/upload` and `/images` are automatically protected by the middleware. Unauthenticated users are redirected to `/auth/login`.

To make additional pages public, update `middleware.ts`:

```typescript
export const config = {
  matcher: [
    // Add routes to exclude from protection
    "/((?!_next/static|_next/image|favicon.ico|public-page).*)",
  ],
};
```

### Accessing User Information

Use the `useUser` hook in client components:

```tsx
"use client";

import { useUser } from "@auth0/nextjs-auth0";

export default function MyComponent() {
  const { user, error, isLoading } = useUser();

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  if (!user) return <div>Not logged in</div>;

  return (
    <div>
      <p>Welcome, {user.name}!</p>
      <p>Email: {user.email}</p>
    </div>
  );
}
```

### Making API Requests

The `apiFetch` function automatically includes the access token:

```typescript
import { apiFetch } from "@/lib/api";

// Access token is automatically included in Authorization header
const projects = await apiFetch<Project[]>("/v1/projects");
```

### Server-Side Access

Access user session in server components or API routes:

```typescript
import { auth0 } from "@/lib/auth0";

export default async function ServerPage() {
  const session = await auth0.getSession();

  if (!session) {
    return <div>Not authenticated</div>;
  }

  return <div>Welcome, {session.user.name}!</div>;
}
```

### Accessing Access Tokens Server-Side

```typescript
import { auth0 } from "@/lib/auth0";

export async function GET(request: NextRequest) {
  const { accessToken } = await auth0.getAccessToken();

  // Use accessToken to call backend API
  const response = await fetch("http://api:8080/v1/protected", {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });

  return Response.json(await response.json());
}
```

## Auth Routes

The Auth0 middleware automatically handles these routes:

| Route                      | Description                                       |
| -------------------------- | ------------------------------------------------- |
| `/auth/login`              | Initiates login flow (redirects to Auth0)         |
| `/auth/logout`             | Logs out user and clears session                  |
| `/auth/callback`           | OAuth callback (Auth0 redirects here after login) |
| `/auth/profile`            | Returns user profile JSON                         |
| `/auth/access-token`       | Returns access token JSON (for client-side use)   |
| `/auth/backchannel-logout` | Handles back-channel logout from Auth0            |

## SSE with Authentication

EventSource doesn't support custom headers, so we pass the access token as a query parameter:

```typescript
const response = await fetch("/auth/access-token");
const { accessToken } = await response.json();
const url = `/api/v1/events?image_id=${imageId}&access_token=${accessToken}`;
const eventSource = new EventSource(url);
```

The backend validates the `access_token` query parameter.

## Security Considerations

### Session Cookie Security

- **httpOnly**: JavaScript cannot access the cookie (XSS protection)
- **Secure**: Cookie only sent over HTTPS in production
- **SameSite=Lax**: CSRF protection
- **Encrypted**: Cookie contents encrypted with `AUTH0_SECRET`

### Token Handling

- **Access Tokens**: Short-lived (1 hour), automatically refreshed
- **Refresh Tokens**: Stored in encrypted session cookie, used to get new access tokens
- **ID Tokens**: User profile information, validated by SDK

### Best Practices

1. **Keep `AUTH0_SECRET` secure**: Use a strong, random secret (≥32 chars)
2. **Rotate secrets regularly**: Follow secret rotation procedures
3. **Use HTTPS in production**: Set `APP_BASE_URL` to `https://` URL
4. **Validate redirect URLs**: Auth0 validates callbacks, but double-check configuration
5. **Monitor failed logins**: Use Auth0 Dashboard to track authentication errors

## Migration from Phase 1

### What Changed

| Phase 1                           | Phase 2 (Current)                     |
| --------------------------------- | ------------------------------------- |
| Manual token entry via `TokenBar` | OAuth flow with Auth0 Universal Login |
| Token stored in `localStorage`    | Token in encrypted, httpOnly cookie   |
| No automatic refresh              | Automatic token refresh               |
| No protected routes               | Middleware protects routes            |
| Manual `make token` command       | Auth0 login UI                        |

### Removed Components

- `TokenBar.tsx` - Replaced by `AuthButton.tsx`
- Manual `localStorage.getItem('token')` - Use `/auth/access-token` endpoint

### Migration Steps for Developers

1. **Install dependencies**: Already done via `npm install @auth0/nextjs-auth0`
2. **Configure Auth0**: Create application in Auth0 Dashboard
3. **Set environment variables**: Copy `env.example` to `.env.local` and fill in values
4. **Test login flow**: Start dev server, click "Login", authenticate with Auth0
5. **Verify API calls**: Check that API requests include valid access tokens

## Troubleshooting

### "Failed to get access token"

- Check that `AUTH0_AUDIENCE` matches the backend API configuration
- Verify Auth0 application has API access enabled

### Redirect Loop

- Ensure Auth0 callback URL matches exactly: `http://localhost:3000/auth/callback`
- Check `APP_BASE_URL` is set correctly

### "Session not found"

- Clear cookies and try logging in again
- Verify `AUTH0_SECRET` is set (≥32 characters)

### API Returns 401 Unauthorized

- Check that backend `AUTH0_DOMAIN` and `AUTH0_AUDIENCE` match frontend
- Verify access token is being sent in `Authorization: Bearer <token>` header
- Use browser DevTools Network tab to inspect request headers

### Token Refresh Failing

- Ensure Auth0 application has "Refresh Token" grant enabled
- Check that `AUTH0_SCOPE` includes `offline_access` if refresh tokens needed

## Testing

### Manual Testing Checklist

- [ ] Click "Login" → redirected to Auth0 Universal Login
- [ ] Enter credentials → redirected back to app
- [ ] User email/name displayed in header
- [ ] API requests include access token (check Network tab)
- [ ] Upload image → presigned URL works
- [ ] SSE viewer connects with access token
- [ ] Click "Logout" → session cleared, redirected to homepage
- [ ] Access `/upload` while logged out → redirected to login

### Auth0 Test Users

Create test users in Auth0 Dashboard → User Management → Users for development testing.

## Development Workflow

```bash
# 1. Start backend API
make up

# 2. Start Next.js dev server
cd apps/web
npm run dev

# 3. Open browser
open http://localhost:3000

# 4. Click "Login" and authenticate with Auth0
```

## Production Deployment

1. **Set environment variables** in your hosting platform (Vercel, AWS, etc.)
2. **Update Auth0 callback URLs** to production domain
3. **Set `APP_BASE_URL`** to production URL (`https://yourdomain.com`)
4. **Enable HTTPS** for secure cookie transmission
5. **Rotate `AUTH0_SECRET`** (don't reuse dev secret)

## References

- [Auth0 Next.js SDK Documentation](https://github.com/auth0/nextjs-auth0)
- [Auth0 Next.js Quickstart](https://auth0.com/docs/quickstart/webapp/nextjs)
- [Auth0 Dashboard](https://manage.auth0.com)
- [OpenID Connect Specification](https://openid.net/specs/openid-connect-core-1_0.html)

## Support

For issues with Auth0 SDK integration:

1. Check this documentation
2. Review Auth0 SDK examples: https://github.com/auth0/nextjs-auth0/blob/main/EXAMPLES.md
3. Check Auth0 logs in Dashboard → Monitoring → Logs
4. Open issue in this repository with steps to reproduce
