import type { NextRequest } from 'next/server';
import { auth0 } from './lib/auth0';

/**
 * Middleware to handle authentication.
 * 
 * Auth0 SDK v4 middleware automatically handles:
 * - /auth/login - Redirects to Auth0 Universal Login
 * - /auth/logout - Logs out and clears session
 * - /auth/callback - Handles Auth0 callback
 * - /auth/profile - Returns user profile
 * - /auth/access-token - Returns access token (if enabled)
 * - Session management and rolling sessions
 * - Protecting routes based on authentication
 */
export async function middleware(request: NextRequest) {
  return await auth0.middleware(request);
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico, sitemap.xml, robots.txt (metadata files)
     */
    '/((?!_next/static|_next/image|favicon.ico|sitemap.xml|robots.txt).*)',
  ],
};
