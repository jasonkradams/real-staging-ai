import type { NextRequest } from 'next/server';
import { NextResponse } from 'next/server';
import { auth0 } from './lib/auth0';

/**
 * Middleware to handle authentication for protected routes.
 * 
 * This middleware runs on the Edge and automatically manages
 * Auth0 session cookies and redirects for authentication flows.
 * 
 * Protected routes: /upload, /images
 * Public routes: /, /health, /api/* (except /api/v1/*)
 */
export async function middleware(request: NextRequest) {
  // Check if Auth0 is configured
  const isAuth0Configured = 
    process.env.AUTH0_DOMAIN && 
    process.env.AUTH0_CLIENT_ID && 
    process.env.APP_BASE_URL && 
    process.env.AUTH0_SECRET;

  if (!isAuth0Configured) {
    // Auth0 not configured - show helpful error message
    const url = request.nextUrl.clone();
    
    // Allow access to the homepage to show configuration instructions
    if (url.pathname === '/') {
      return NextResponse.next();
    }
    
    // For auth routes and protected routes, show configuration error
    return new NextResponse(
      JSON.stringify({
        error: 'Auth0 Not Configured',
        message: 'Please configure Auth0 environment variables in .env.local',
        required: [
          'AUTH0_DOMAIN',
          'AUTH0_CLIENT_ID',
          'AUTH0_CLIENT_SECRET',
          'AUTH0_SECRET',
          'APP_BASE_URL'
        ],
        instructions: 'See apps/web/env.example for configuration template'
      }),
      {
        status: 503,
        headers: {
          'content-type': 'application/json',
        },
      }
    );
  }

  // Auth0 is configured - run middleware
  try {
    return await auth0.middleware(request);
  } catch (error) {
    console.error('Auth0 middleware error:', error);
    return new NextResponse(
      JSON.stringify({
        error: 'Authentication Error',
        message: error instanceof Error ? error.message : 'Unknown error',
      }),
      {
        status: 500,
        headers: {
          'content-type': 'application/json',
        },
      }
    );
  }
}

export const config = {
  matcher: [
    /*
     * Only match protected routes and auth routes.
     * This allows the homepage (/) to be publicly accessible.
     */
    '/upload/:path*',
    '/images/:path*',
    '/auth/:path*',
  ],
};
