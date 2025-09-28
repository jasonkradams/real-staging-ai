import './globals.css'
import Link from 'next/link'
import type { Metadata } from 'next'
import TokenBar from '@/components/TokenBar'

export const metadata: Metadata = {
  title: 'Virtual Staging AI',
  description: 'Phase 1 Web UI',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <header className="border-b bg-white">
          <nav className="container flex items-center justify-between py-3">
            <div className="flex items-center gap-6">
              <Link href="/" className="font-semibold">Virtual Staging AI</Link>
              <Link href="/upload" className="text-gray-600 hover:text-gray-900">Upload</Link>
              <Link href="/images" className="text-gray-600 hover:text-gray-900">Images</Link>
            </div>
            <div className="flex items-center gap-4">
              <TokenBar />
              <div className="text-sm text-gray-500">Phase 1</div>
            </div>
          </nav>
        </header>
        <main className="container py-6">
          {children}
        </main>
      </body>
    </html>
  )
}
