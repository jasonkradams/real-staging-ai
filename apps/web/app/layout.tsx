import './globals.css'
import Link from 'next/link'
import type { Metadata } from 'next'
import { Sparkles } from 'lucide-react'
import AuthButton from '@/components/AuthButton'
import UserProvider from '@/components/UserProvider'
import ProtectedNav from '@/components/ProtectedNav'
import { ThemeProvider } from '@/components/ThemeProvider'
import { ThemeToggle } from '@/components/ThemeToggle'

export const metadata: Metadata = {
  title: 'Real Staging AI | Transform Properties with AI',
  description: 'Professional AI-powered virtual staging for real estate. Transform empty spaces into beautifully furnished rooms in seconds.',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="h-full" suppressHydrationWarning>
      <body className="h-full">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <UserProvider>
          <div className="flex min-h-screen flex-col">
            {/* Gradient Header */}
            <header className="sticky top-0 z-50 w-full border-b border-gray-200/60 dark:border-gray-800/60 bg-white/80 dark:bg-slate-950/80 backdrop-blur-xl supports-[backdrop-filter]:bg-white/60 dark:supports-[backdrop-filter]:bg-slate-950/60">
              <nav className="container flex h-16 items-center justify-between">
                <div className="flex items-center gap-8">
                  <Link href="/" className="flex items-center gap-2 font-bold text-lg group">
                    <div className="rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 p-2 shadow-lg shadow-blue-500/30 transition-all group-hover:shadow-xl group-hover:shadow-blue-500/40">
                      <Sparkles className="h-5 w-5 text-white" />
                    </div>
                    <span className="gradient-text hidden sm:inline">Real Staging AI</span>
                  </Link>
                  <div className="hidden items-center gap-1 md:flex">
                    <ProtectedNav />
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <ThemeToggle />
                  <AuthButton />
                </div>
              </nav>
            </header>

            {/* Main Content */}
            <main className="flex-1">
              <div className="container py-8 lg:py-12 animate-in">
                {children}
              </div>
            </main>

            {/* Footer */}
            <footer className="border-t border-gray-200/60 dark:border-gray-800/60 bg-white/80 dark:bg-slate-950/80 backdrop-blur-sm">
              <div className="container py-6">
                <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    © {new Date().getFullYear()} Real Staging AI. Built with Next.js & Replicate.
                  </p>
                  <div className="flex gap-4">
                    {/* Protected links removed from footer - only available when authenticated */}
                  </div>
                </div>
              </div>
            </footer>
          </div>
          </UserProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
