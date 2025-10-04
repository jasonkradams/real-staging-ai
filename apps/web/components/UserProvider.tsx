'use client';

import { Auth0Provider } from '@auth0/nextjs-auth0';

/**
 * UserProvider wraps the Auth0Provider to provide user session context
 * throughout the application. This must be a client component.
 */
export default function UserProvider({ children }: { children: React.ReactNode }) {
  return <Auth0Provider>{children}</Auth0Provider>;
}
