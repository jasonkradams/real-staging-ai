import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import React from 'react'
import { createRoot } from 'react-dom/client'
import { act } from 'react-dom/test-utils'

// Mock Auth0 hook
vi.mock('@auth0/nextjs-auth0', () => ({
  useUser: vi.fn(() => ({
    user: { email: 'user@example.com', name: 'Auth User' },
    error: null,
    isLoading: false,
  })),
}))

// Mock next/link to a basic anchor for jsdom
vi.mock('next/link', () => ({
  default: (props: React.AnchorHTMLAttributes<HTMLAnchorElement> & { href?: string; children?: React.ReactNode }) => React.createElement('a', props),
}))

// Mock api client
vi.mock('@/lib/api', () => ({
  apiFetch: vi.fn(async () => ({
    id: 'id',
    role: 'user',
    created_at: '',
    updated_at: '',
    full_name: 'Profile Name',
  })),
}))

import AuthButton from './AuthButton'
import { apiFetch } from '@/lib/api'
import type { Mock } from 'vitest'

function render(ui: React.ReactElement) {
  const container = document.createElement('div')
  document.body.appendChild(container)
  const root = createRoot(container)
  act(() => {
    root.render(ui)
  })
  return { container, root }
}

describe('AuthButton', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('success: shows profile full_name from backend when available', async () => {
    const { container } = render(<AuthButton />)

    // Wait for effect to fetch and update state
    await act(async () => {
      await Promise.resolve()
    })

    expect(apiFetch).toHaveBeenCalledWith('/v1/user/profile')

    // Expect the display name to include backend profile name
    expect(container.textContent || '').toContain('Profile Name')
  })

  it('success: falls back to Auth0 name when profile full_name is not set', async () => {
    ;(apiFetch as unknown as Mock).mockResolvedValueOnce({
      id: 'id',
      role: 'user',
      created_at: '',
      updated_at: '',
    })

    const { container } = render(<AuthButton />)

    await act(async () => {
      await Promise.resolve()
    })

    expect(container.textContent || '').toContain('Auth User')
  })
})
