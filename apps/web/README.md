# Real Staging AI - Web Frontend

Next.js web application for Real Staging AI with Auth0 authentication.

## Tech Stack

- **Framework**: Next.js 15.5.4 (App Router)
- **Language**: TypeScript 5.6.2
- **Styling**: Tailwind CSS 3.4.13
- **Authentication**: Auth0 (@auth0/nextjs-auth0 v4)
- **Node**: 18.18.0+

## Quick Start

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Environment Variables

```bash
# Copy the example environment file
cp env.example .env.local

# Edit .env.local and fill in your Auth0 credentials
```

Required environment variables:
- `AUTH0_DOMAIN` - Your Auth0 tenant domain
- `AUTH0_CLIENT_ID` - Auth0 application client ID
- `AUTH0_CLIENT_SECRET` - Auth0 application client secret
- `AUTH0_SECRET` - Secret for encrypting sessions (generate with `openssl rand -hex 32`)
- `APP_BASE_URL` - Application URL (`http://localhost:3000` for dev)
- `AUTH0_AUDIENCE` - Auth0 API audience (must match backend: `https://api.realstaging.local`)

See `env.example` for complete configuration.

### 3. Configure Auth0 Application

In the [Auth0 Dashboard](https://manage.auth0.com):

1. Create a **Regular Web Application**
2. Add to **Allowed Callback URLs**: `http://localhost:3000/auth/callback`
3. Add to **Allowed Logout URLs**: `http://localhost:3000`
4. Add to **Allowed Web Origins**: `http://localhost:3000`

### 4. Start Backend API

The frontend requires the backend API to be running:

```bash
# In repository root
make up
```

This starts the API on `http://localhost:8080`.

### 5. Run Development Server

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

## Features

### Authentication (Auth0)
- OAuth 2.0 / OpenID Connect flow
- Universal Login with Auth0 hosted page
- Automatic token refresh
- Secure session management (httpOnly, encrypted cookies)
- Protected routes with automatic redirect

### Pages
- **Dashboard** (`/`) - Landing page with navigation
- **Upload** (`/upload`) - Upload images to projects (protected)
- **Images** (`/images`) - View and manage images (protected)

### Core Functionality
- **Project Management**: Create and select projects
- **Image Upload**: Presigned S3 upload flow
- **Real-time Updates**: SSE client for job status updates
- **Image Viewing**: Presigned URLs for original and staged images

## Project Structure

```
apps/web/
├── app/
│   ├── layout.tsx          # Root layout with Auth0Provider
│   ├── page.tsx            # Dashboard
│   ├── upload/
│   │   └── page.tsx        # Upload page (protected)
│   ├── images/
│   │   └── page.tsx        # Images page (protected)
│   └── globals.css         # Global styles
├── components/
│   ├── UserProvider.tsx    # Auth0 provider wrapper
│   ├── AuthButton.tsx      # Login/logout UI
│   └── SSEViewer.tsx       # SSE client for job updates
├── lib/
│   ├── auth0.ts           # Auth0 SDK client
│   └── api.ts             # API client with automatic auth
├── middleware.ts          # Auth0 session middleware
├── next.config.js         # Next.js config with API proxy
└── env.example            # Environment variable template
```

## Development

### Running Tests

```bash
npm run lint        # Run ESLint
npm run build       # Test production build
```

### API Proxy

The Next.js app proxies `/api/*` requests to the backend API at `http://localhost:8080/api/*`. This is configured in `next.config.js`.

### Authentication Flow

1. User visits protected route → middleware redirects to `/auth/login`
2. Auth0 Universal Login → user authenticates
3. Callback to `/auth/callback` → middleware creates session
4. User redirected to originally requested page
5. Access token automatically included in API requests

## Documentation

- **Auth0 Integration**: [docs/frontend/AUTH0_INTEGRATION.md](../../docs/frontend/AUTH0_INTEGRATION.md)
- **Phase 1 Implementation**: [docs/frontend/PHASE1_IMPLEMENTATION.md](../../docs/frontend/PHASE1_IMPLEMENTATION.md)
- **API Documentation**: https://jasonkradams.github.io/real-staging-ai/

## Troubleshooting

### "Failed to get access token"
- Verify `AUTH0_AUDIENCE` matches backend configuration
- Check Auth0 application has API access enabled

### Redirect loop
- Ensure Auth0 callback URL is exactly: `http://localhost:3000/auth/callback`
- Verify `APP_BASE_URL` is set correctly

### API returns 401
- Check backend `AUTH0_DOMAIN` and `AUTH0_AUDIENCE` match frontend
- Verify backend API is running (`make up`)
- Check Network tab in browser DevTools for request headers

### "Session not found"
- Clear cookies and log in again
- Verify `AUTH0_SECRET` is set (≥32 characters)

See [AUTH0_INTEGRATION.md](../../docs/frontend/AUTH0_INTEGRATION.md) for more troubleshooting tips.

## Production Build

```bash
npm run build       # Create production build
npm start           # Start production server
```

**Important for production:**
- Set `APP_BASE_URL` to production domain (HTTPS)
- Update Auth0 callback URLs to production domain
- Use secure, unique `AUTH0_SECRET` (don't reuse dev secret)
- Enable proper CORS settings on backend API

## Contributing

Follow the repository guidelines in [AGENTS.md](../../AGENTS.md) and use Conventional Commits for all changes.
