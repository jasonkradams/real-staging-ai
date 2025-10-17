'use client';

import { useUser } from '@auth0/nextjs-auth0';
import { LogIn, LogOut, Loader2, User as UserIcon, ChevronDown } from 'lucide-react';
import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { apiFetch } from '@/lib/api';
import type { BackendProfile } from '@/lib/profile';

/**
 * AuthButton displays login/logout button based on user session state
 * When logged in, shows user avatar/name with dropdown menu
 */
export default function AuthButton() {
  const { user, error, isLoading } = useUser();
  const [profileName, setProfileName] = useState<string | null>(null);
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Fetch user's configured profile name
  useEffect(() => {
    if (user) {
      apiFetch<BackendProfile>('/v1/user/profile')
        .then((data) => {
          if (data?.full_name) {
            setProfileName(data.full_name);
          }
        })
        .catch(() => {
          // Silently fail - will use Auth0 name
        });
    }
  }, [user]);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownOpen(false);
      }
    }

    if (dropdownOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [dropdownOpen]);

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400">
        <Loader2 className="h-4 w-4 animate-spin" />
        <span className="text-sm hidden sm:inline">Loading...</span>
      </div>
    );
  }

  // Hide "unauthorized" error (expected when not logged in)
  // Only show actual errors
  if (error && !error.message?.toLowerCase().includes('unauthorized')) {
    return (
      <div className="flex items-center gap-2 rounded-lg bg-red-50 dark:bg-red-950/20 px-3 py-2 text-sm text-red-700 dark:text-red-400">
        <span>⚠️ Auth error</span>
      </div>
    );
  }

  if (user) {
    // Use profile name if configured, otherwise fallback to Auth0 name or email
    const displayName = profileName || user.name || user.email?.split('@')[0] || 'User';
    const userInitial = displayName.charAt(0).toUpperCase();

    return (
      <div className="relative" ref={dropdownRef}>
        {/* User Avatar/Name Button */}
        <button
          onClick={() => setDropdownOpen(!dropdownOpen)}
          className="flex items-center gap-2 rounded-lg px-3 py-2 transition-colors hover:bg-gray-100 dark:hover:bg-slate-800"
        >
          {/* Avatar */}
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-indigo-500 text-sm font-semibold text-white shadow-lg shadow-blue-500/30">
            {userInitial}
          </div>
          
          {/* Name (hidden on mobile) */}
          <div className="hidden items-start sm:flex flex-col text-left">
            <div className="text-sm font-medium text-gray-900 dark:text-gray-100">
              {displayName}
            </div>
            {user.email && profileName && (
              <div className="text-xs text-gray-500 dark:text-gray-400">{user.email}</div>
            )}
          </div>

          {/* Dropdown Arrow */}
          <ChevronDown className={`h-4 w-4 text-gray-500 transition-transform hidden sm:block ${dropdownOpen ? 'rotate-180' : ''}`} />
        </button>

        {/* Dropdown Menu */}
        {dropdownOpen && (
          <div className="absolute right-0 mt-2 w-56 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-slate-900 shadow-lg z-50">
            <div className="p-2">
              {/* User info (mobile only) */}
              <div className="sm:hidden px-3 py-2 border-b border-gray-200 dark:border-gray-800 mb-2">
                <div className="text-sm font-medium text-gray-900 dark:text-gray-100">
                  {displayName}
                </div>
                {user.email && (
                  <div className="text-xs text-gray-500 dark:text-gray-400">{user.email}</div>
                )}
              </div>

              {/* Menu Items */}
              <Link
                href="/profile"
                onClick={() => setDropdownOpen(false)}
                className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 transition-colors hover:bg-gray-100 dark:hover:bg-slate-800"
              >
                <UserIcon className="h-4 w-4" />
                Profile Settings
              </Link>

              <a
                href="/auth/logout"
                className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-red-600 dark:text-red-400 transition-colors hover:bg-red-50 dark:hover:bg-red-950/20"
              >
                <LogOut className="h-4 w-4" />
                Logout
              </a>
            </div>
          </div>
        )}
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
