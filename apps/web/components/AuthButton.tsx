'use client';

import { useUser } from '@auth0/nextjs-auth0';

/**
 * AuthButton displays login/logout button based on user session state
 */
export default function AuthButton() {
  const { user, error, isLoading } = useUser();

  if (isLoading) {
    return <div className="text-sm text-gray-500">Loading...</div>;
  }

  if (error) {
    return <div className="text-sm text-red-500">Error: {error.message}</div>;
  }

  if (user) {
    return (
      <div className="flex items-center gap-4">
        <div className="text-sm text-gray-700">
          {user.email || user.name || 'User'}
        </div>
        <a
          href="/api/auth/logout"
          className="text-sm text-gray-600 hover:text-gray-900 underline"
        >
          Logout
        </a>
      </div>
    );
  }

  return (
    <a
      href="/api/auth/login"
      className="text-sm text-blue-600 hover:text-blue-800 underline"
    >
      Login
    </a>
  );
}
