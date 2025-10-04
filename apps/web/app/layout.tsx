import './globals.css'
import Link from 'next/link'
import type { Metadata } from 'next'
import AuthButton from '@/components/AuthButton'
import UserProvider from '@/components/UserProvider'

export const metadata: Metadata = {
  title: 'Virtual Staging AI',
  description: 'Virtual Staging AI - Phase 2 with Auth0',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <UserProvider>
          <header className="border-b bg-white">
            <nav className="container flex items-center justify-between py-3">
              <div className="flex items-center gap-6">
                <Link href="/" className="font-semibold">Virtual Staging AI</Link>
                <Link href="/upload" className="text-gray-600 hover:text-gray-900">Upload</Link>
                <Link href="/images" className="text-gray-600 hover:text-gray-900">Images</Link>
              </div>
              <AuthButton />
            </nav>
          </header>
          <main className="container py-6">
            {children}
          </main>
        </UserProvider>
      </body>
    </html>
  )
}
