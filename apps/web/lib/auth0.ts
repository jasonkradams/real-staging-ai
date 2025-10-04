import { Auth0Client } from '@auth0/nextjs-auth0/server';

/**
 * Auth0 SDK Client
 * 
 * CRITICAL: You MUST set AUTH0_AUDIENCE in .env.local to get API access tokens!
 * Without it, Auth0 returns encrypted ID tokens instead of JWT access tokens.
 * 
 * Required environment variables:
 * - AUTH0_SECRET: Long random string for encrypting the session cookie (>= 32 chars)
 * - APP_BASE_URL: The base URL of your application (e.g., http://localhost:3000)
 * - AUTH0_DOMAIN: Your Auth0 domain (e.g., dev-sleeping-pandas.us.auth0.com)
 * - AUTH0_CLIENT_ID: Your Auth0 application client ID
 * - AUTH0_CLIENT_SECRET: Your Auth0 application client secret
 * - AUTH0_AUDIENCE: Your Auth0 API audience (e.g., https://api.virtualstaging.local) - REQUIRED!
 * 
 * Optional environment variables:
 * - AUTH0_SCOPE: OAuth scopes (default: openid profile email offline_access)
 */
export const auth0 = new Auth0Client({
  // Enable the /auth/access-token endpoint for client-side API calls
  enableAccessTokenEndpoint: true,
  
  // Ensure we request the audience if not set via env var
  authorizationParameters: {
    audience: process.env.AUTH0_AUDIENCE,
    scope: process.env.AUTH0_SCOPE || 'openid profile email offline_access',
  },
});
