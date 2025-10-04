'use client';

import { useUser } from '@auth0/nextjs-auth0';

export default function Page() {
  const { user, isLoading, error } = useUser();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-semibold">Welcome to Virtual Staging AI</h1>
        <p className="text-gray-600">Loading...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-semibold">Welcome to Virtual Staging AI</h1>
        <p className="text-gray-600">
          Transform your property photos with AI-powered virtual staging.
        </p>

        {error && error.message?.includes('failed to fetch') ? (
          <div className="card bg-red-50 border-red-200">
            <div className="card-body">
              <h2 className="text-lg font-semibold mb-2 text-red-800">⚠️ Auth0 Not Configured</h2>
              <p className="text-gray-700 mb-4">
                Please configure Auth0 environment variables in <code className="bg-red-100 px-2 py-1 rounded">.env.local</code>
              </p>
              <div className="text-sm text-gray-600 space-y-2">
                <p><strong>Required variables:</strong></p>
                <ul className="list-disc list-inside pl-4 space-y-1">
                  <li><code>AUTH0_DOMAIN</code></li>
                  <li><code>AUTH0_CLIENT_ID</code></li>
                  <li><code>AUTH0_CLIENT_SECRET</code></li>
                  <li><code>AUTH0_SECRET</code> (generate with: <code>openssl rand -hex 32</code>)</li>
                  <li><code>APP_BASE_URL=http://localhost:3000</code></li>
                </ul>
                <p className="mt-4">
                  See <code className="bg-red-100 px-2 py-1 rounded">apps/web/env.example</code> for the complete template.
                </p>
              </div>
            </div>
          </div>
        ) : (
          <div className="card bg-blue-50 border-blue-200">
            <div className="card-body">
              <h2 className="text-lg font-semibold mb-2">Get Started</h2>
              <p className="text-gray-700 mb-4">
                Sign in to upload images and start staging your properties.
              </p>
              <a
                href="/auth/login"
                className="inline-block bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700 transition-colors"
              >
                Sign In
              </a>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="card">
            <div className="card-header">Upload Images</div>
            <div className="card-body text-gray-600">
              Upload property photos directly to secure cloud storage.
            </div>
          </div>
          <div className="card">
            <div className="card-header">Real-time Processing</div>
            <div className="card-body text-gray-600">
              Track image processing status with live updates.
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Welcome back, {user.name || user.email}!</h1>
      <p className="text-gray-600">Upload images and track processing status in real-time.</p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <a href="/upload" className="card hover:shadow-lg transition-shadow">
          <div className="card-header">Upload</div>
          <div className="card-body text-gray-600">
            Presign, upload to S3, and create an image job.
          </div>
        </a>
        <a href="/images" className="card hover:shadow-lg transition-shadow">
          <div className="card-header">Images</div>
          <div className="card-body text-gray-600">
            View your images and track processing status in real-time.
          </div>
        </a>
      </div>
    </div>
  );
}
