'use client';

import { useUser } from '@auth0/nextjs-auth0';
import Link from 'next/link';
import { Sparkles, Upload, ImageIcon, Zap, Shield, Clock, ArrowRight, AlertCircle } from 'lucide-react';

export default function Page() {
  const { user, isLoading, error } = useUser();

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-4">
        <div className="relative">
          <div className="h-16 w-16 rounded-full border-4 border-gray-200 border-t-blue-600 animate-spin"></div>
          <Sparkles className="absolute inset-0 m-auto h-6 w-6 text-blue-600" />
        </div>
        <p className="text-gray-600 animate-pulse">Loading your workspace...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="space-y-16">
        {/* Hero Section */}
        <section className="relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-r from-blue-50 to-indigo-50 opacity-50 -z-10"></div>
          <div className="mx-auto max-w-4xl text-center space-y-6 py-12">
            <div className="inline-flex items-center gap-2 rounded-full bg-blue-100 px-4 py-2 text-sm font-medium text-blue-700">
              <Sparkles className="h-4 w-4" />
              AI-Powered Virtual Staging
            </div>
            <h1 className="text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl">
              Transform Empty Spaces into
              <span className="gradient-text block">Stunning Showcase Homes</span>
            </h1>
            <p className="mx-auto max-w-2xl text-lg text-gray-600">
              Professional virtual staging powered by cutting-edge AI. Turn vacant properties 
              into beautifully furnished spaces in seconds, not days.
            </p>

            {error && error.message?.includes('failed to fetch') ? (
              <div className="card mx-auto max-w-2xl border-red-200 bg-red-50/50">
                <div className="card-body space-y-4">
                  <div className="flex items-start gap-3">
                    <AlertCircle className="h-5 w-5 text-red-600 mt-0.5" />
                    <div className="flex-1 text-left">
                      <h2 className="text-lg font-semibold text-red-800">Auth0 Configuration Required</h2>
                      <p className="text-sm text-gray-700 mt-2">
                        Please configure Auth0 environment variables in <code className="bg-red-100 px-2 py-1 rounded text-xs">.env.local</code>
                      </p>
                      <div className="mt-4 space-y-2 text-sm text-gray-600">
                        <p className="font-medium">Required variables:</p>
                        <ul className="list-disc list-inside pl-4 space-y-1 text-xs">
                          <li><code className="bg-red-50 px-1 py-0.5 rounded">AUTH0_DOMAIN</code></li>
                          <li><code className="bg-red-50 px-1 py-0.5 rounded">AUTH0_CLIENT_ID</code></li>
                          <li><code className="bg-red-50 px-1 py-0.5 rounded">AUTH0_CLIENT_SECRET</code></li>
                          <li><code className="bg-red-50 px-1 py-0.5 rounded">AUTH0_SECRET</code> (generate with: <code className="text-xs">openssl rand -hex 32</code>)</li>
                          <li><code className="bg-red-50 px-1 py-0.5 rounded">APP_BASE_URL=http://localhost:3000</code></li>
                        </ul>
                        <p className="mt-3">
                          See <code className="bg-red-100 px-1 py-0.5 rounded text-xs">apps/web/env.example</code> for details.
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
                <a href="/auth/login" className="btn btn-primary group">
                  Get Started
                  <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
                </a>
                <a href="#features" className="btn btn-secondary">
                  Learn More
                </a>
              </div>
            )}
          </div>
        </section>

        {/* Features Section */}
        <section id="features" className="space-y-8">
          <div className="text-center space-y-3">
            <h2 className="text-3xl font-bold">Why Choose Virtual Staging AI?</h2>
            <p className="text-gray-600 max-w-2xl mx-auto">
              Professional results, instant delivery, and a seamless workflow designed for real estate professionals.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div className="card group cursor-default">
              <div className="card-body space-y-4">
                <div className="rounded-xl bg-gradient-to-br from-blue-500 to-indigo-500 p-3 w-fit shadow-lg shadow-blue-500/30 transition-all group-hover:shadow-xl group-hover:shadow-blue-500/40">
                  <Upload className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-lg mb-2">Direct Cloud Upload</h3>
                  <p className="text-gray-600 text-sm">
                    Securely upload property photos directly to cloud storage with pre-signed URLs. No server bottlenecks.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default">
              <div className="card-body space-y-4">
                <div className="rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 p-3 w-fit shadow-lg shadow-purple-500/30 transition-all group-hover:shadow-xl group-hover:shadow-purple-500/40">
                  <Zap className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-lg mb-2">Lightning Fast</h3>
                  <p className="text-gray-600 text-sm">
                    AI-powered processing delivers professional staging results in seconds. Watch real-time progress updates.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default">
              <div className="card-body space-y-4">
                <div className="rounded-xl bg-gradient-to-br from-green-500 to-emerald-500 p-3 w-fit shadow-lg shadow-green-500/30 transition-all group-hover:shadow-xl group-hover:shadow-green-500/40">
                  <Clock className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-lg mb-2">Real-time Tracking</h3>
                  <p className="text-gray-600 text-sm">
                    Monitor job status with live Server-Sent Events. Know exactly when your staged images are ready.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default">
              <div className="card-body space-y-4">
                <div className="rounded-xl bg-gradient-to-br from-orange-500 to-red-500 p-3 w-fit shadow-lg shadow-orange-500/30 transition-all group-hover:shadow-xl group-hover:shadow-orange-500/40">
                  <ImageIcon className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-lg mb-2">Style Customization</h3>
                  <p className="text-gray-600 text-sm">
                    Choose from multiple styles and room types. Modern, traditional, minimalist — we&apos;ve got you covered.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default">
              <div className="card-body space-y-4">
                <div className="rounded-xl bg-gradient-to-br from-cyan-500 to-blue-500 p-3 w-fit shadow-lg shadow-cyan-500/30 transition-all group-hover:shadow-xl group-hover:shadow-cyan-500/40">
                  <Shield className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-lg mb-2">Secure & Private</h3>
                  <p className="text-gray-600 text-sm">
                    Enterprise-grade security with Auth0 authentication and encrypted storage. Your data stays protected.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default">
              <div className="card-body space-y-4">
                <div className="rounded-xl bg-gradient-to-br from-violet-500 to-purple-500 p-3 w-fit shadow-lg shadow-violet-500/30 transition-all group-hover:shadow-xl group-hover:shadow-violet-500/40">
                  <Sparkles className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-lg mb-2">AI-Powered</h3>
                  <p className="text-gray-600 text-sm">
                    Leveraging Replicate&apos;s cutting-edge models to deliver photorealistic staging that sells properties.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    );
  }

  return (
    <div className="space-y-12">
      {/* Welcome Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-r from-blue-50 to-indigo-50 opacity-50 rounded-3xl -z-10"></div>
        <div className="px-8 py-12 text-center space-y-4">
          <div className="inline-flex items-center gap-2 rounded-full bg-gradient-to-r from-blue-600 to-indigo-600 px-4 py-1.5 text-sm font-medium text-white shadow-lg shadow-blue-500/30">
            <Sparkles className="h-4 w-4" />
            Dashboard
          </div>
          <h1 className="text-3xl sm:text-4xl font-bold">
            Welcome back, <span className="gradient-text">{user.name || user.email?.split('@')[0]}</span>!
          </h1>
          <p className="text-gray-600 max-w-2xl mx-auto">
            Start staging properties or manage your existing projects. Everything you need is right at your fingertips.
          </p>
        </div>
      </section>

      {/* Quick Actions */}
      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Link href="/upload" className="card group">
          <div className="card-body space-y-4">
            <div className="flex items-start justify-between">
              <div className="rounded-xl bg-gradient-to-br from-blue-500 to-indigo-500 p-3 shadow-lg shadow-blue-500/30 transition-all group-hover:shadow-xl group-hover:shadow-blue-500/40">
                <Upload className="h-6 w-6 text-white" />
              </div>
              <ArrowRight className="h-5 w-5 text-gray-400 transition-all group-hover:translate-x-1 group-hover:text-blue-600" />
            </div>
            <div>
              <h3 className="font-semibold text-xl mb-2 group-hover:text-blue-600 transition-colors">Upload & Stage</h3>
              <p className="text-gray-600 text-sm">
                Upload new property photos and start the AI staging process. Choose room types and styles to match your vision.
              </p>
            </div>
          </div>
        </Link>

        <Link href="/images" className="card group">
          <div className="card-body space-y-4">
            <div className="flex items-start justify-between">
              <div className="rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 p-3 shadow-lg shadow-purple-500/30 transition-all group-hover:shadow-xl group-hover:shadow-purple-500/40">
                <ImageIcon className="h-6 w-6 text-white" />
              </div>
              <ArrowRight className="h-5 w-5 text-gray-400 transition-all group-hover:translate-x-1 group-hover:text-purple-600" />
            </div>
            <div>
              <h3 className="font-semibold text-xl mb-2 group-hover:text-purple-600 transition-colors">View Images</h3>
              <p className="text-gray-600 text-sm">
                Browse your staged images by project. Monitor processing status and download results with live updates.
              </p>
            </div>
          </div>
        </Link>
      </section>
    </div>
  );
}
