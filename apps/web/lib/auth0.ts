import { Auth0Client } from '@auth0/nextjs-auth0/server';

/**
 * Auth0 SDK Client
 * 
 * The Auth0 SDK is automatically configured via environment variables.
 * 
 * Required environment variables:
 * - AUTH0_SECRET: Long random string for encrypting the session cookie (>= 32 chars)
 * - APP_BASE_URL: The base URL of your application (e.g., http://localhost:3000)
 * - AUTH0_DOMAIN: Your Auth0 domain (e.g., dev-sleeping-pandas.us.auth0.com)
 * - AUTH0_CLIENT_ID: Your Auth0 application client ID
 * - AUTH0_CLIENT_SECRET: Your Auth0 application client secret
 * 
 * Optional environment variables:
 * - AUTH0_AUDIENCE: Your Auth0 API audience (e.g., https://api.virtualstaging.local)
 * - AUTH0_SCOPE: OAuth scopes (default: openid profile email)
 */
export const auth0 = new Auth0Client();
