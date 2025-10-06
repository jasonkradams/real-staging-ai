'use client';

import { useUser } from '@auth0/nextjs-auth0';
import { LogIn, LogOut, Loader2 } from 'lucide-react';

/**
 * AuthButton displays login/logout button based on user session state
 */
export default function AuthButton() {
  const { user, error, isLoading } = useUser();

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-gray-500">
        <Loader2 className="h-4 w-4 animate-spin" />
        <span className="text-sm">Loading...</span>
      </div>
    );
  }

  // Hide "unauthorized" error (expected when not logged in)
  // Only show actual errors
  if (error && !error.message?.toLowerCase().includes('unauthorized')) {
    return (
      <div className="flex items-center gap-2 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
        <span>⚠️ Auth error</span>
      </div>
    );
  }

  if (user) {
    const displayName = user.name || user.email?.split('@')[0] || 'User';
    const userInitial = displayName.charAt(0).toUpperCase();

    return (
      <div className="flex items-center gap-3">
        {/* User Info */}
        <div className="hidden items-center gap-3 sm:flex">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-indigo-500 text-sm font-semibold text-white shadow-lg shadow-blue-500/30">
            {userInitial}
          </div>
          <div className="text-right">
            <div className="text-sm font-medium text-gray-900">{displayName}</div>
            {user.email && user.name && (
              <div className="text-xs text-gray-500">{user.email}</div>
            )}
          </div>
        </div>

        {/* Mobile: Just Avatar */}
        <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-indigo-500 text-sm font-semibold text-white shadow-lg shadow-blue-500/30 sm:hidden">
          {userInitial}
        </div>

        {/* Logout Button */}
        <a
          href="/auth/logout"
          className="btn btn-ghost"
        >
          <LogOut className="h-4 w-4" />
          <span className="hidden sm:inline">Logout</span>
        </a>
      </div>
    );
  }

  return (
    <a
      href="/auth/login"
      className="btn btn-primary"
    >
      <LogIn className="h-4 w-4" />
      Sign In
    </a>
  );
}
