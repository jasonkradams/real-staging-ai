'use client';

import { useUser } from '@auth0/nextjs-auth0';
import Link from 'next/link';
import { Upload, ImageIcon } from 'lucide-react';

/**
 * ProtectedNav renders navigation links that are only visible to authenticated users
 */
export default function ProtectedNav() {
  const { user, isLoading } = useUser();

  if (isLoading || !user) return null;

  return (
    <>
      <Link
        href="/upload"
        className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors hover:bg-gray-100 dark:hover:bg-slate-800"
      >
        <Upload className="h-4 w-4" />
        Upload
      </Link>
      <Link
        href="/images"
        className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors hover:bg-gray-100 dark:hover:bg-slate-800"
      >
        <ImageIcon className="h-4 w-4" />
        Images
      </Link>
    </>
  );
}
