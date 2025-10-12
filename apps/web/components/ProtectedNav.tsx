'use client';

import { useUser } from '@auth0/nextjs-auth0';
import Link from 'next/link';
import { Upload, ImageIcon } from 'lucide-react';

/**
 * ProtectedNav renders navigation links that are only visible to authenticated users
 */
export default function ProtectedNav() {
  const { user } = useUser();

  // Don't render anything if user is not authenticated
  if (!user) {
    return null;
  }

  return (
    <div className="flex items-center gap-1">
      <Link href="/upload" className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 transition-colors hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-100">
        <Upload className="h-4 w-4" />
        Upload
      </Link>
      <Link href="/images" className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 transition-colors hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-100">
        <ImageIcon className="h-4 w-4" />
        Images
      </Link>
    </div>
  );
}
