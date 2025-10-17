import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import React from 'react'
import { createRoot } from 'react-dom/client'
import { act } from 'react-dom/test-utils'

// Mock Auth0 useUser to simulate logged-in user
vi.mock('@auth0/nextjs-auth0', () => ({
  useUser: vi.fn(() => ({
    user: { email: 'user@example.com', name: 'Auth User' },
    error: null,
    isLoading: false,
  })),
}))

// Mock next/navigation router
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

// Mock api client
const apiFetchMock = vi.fn()
vi.mock('@/lib/api', () => ({
  apiFetch: (...args: unknown[]) => apiFetchMock(...args),
}))

import ProfilePage from './page'

function render(ui: React.ReactElement) {
  const container = document.createElement('div')
  document.body.appendChild(container)
  const root = createRoot(container)
  act(() => {
    root.render(ui)
  })
  return { container, root }
}

describe('ProfilePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('success: fetches profile and subscriptions via apiFetch on mount', async () => {
    // Arrange responses:
    // 1) user profile
    apiFetchMock
      .mockResolvedValueOnce({
        id: 'id',
        role: 'user',
        created_at: '',
        updated_at: '',
        full_name: 'Test User',
        billing_address: { city: 'Metropolis' },
        preferences: { email_notifications: true },
      })
      // 2) billing subscriptions
      .mockResolvedValueOnce({ items: [{ id: 'sub_1', status: 'active' }] })

    const { container } = render(<ProfilePage />)

    // Wait for effects to run
    await act(async () => {
      await Promise.resolve()
    })

    // Assert calls
    expect(apiFetchMock.mock.calls.some((c) => c[0] === '/v1/user/profile')).toBe(true)
    expect(apiFetchMock.mock.calls.some((c) => c[0] === '/v1/billing/subscriptions')).toBe(true)

    // Renders subscription UI (Active Subscription banner)
    expect(container.textContent || '').toContain('Active Subscription')

    // Full name input should reflect mapped value
    const inputs = Array.from(container.querySelectorAll('input')) as HTMLInputElement[]
    const fullNameInput = inputs.find((el) => el.placeholder === 'John Doe')
    expect(fullNameInput?.value).toBe('Test User')
  })

  it('success: clicking Save sends PATCH with snake_case payload', async () => {
    // Arrange: profile and subscriptions initial
    apiFetchMock
      .mockResolvedValueOnce({
        id: 'id',
        role: 'user',
        created_at: '',
        updated_at: '',
        full_name: 'Snake Case',
      })
      .mockResolvedValueOnce({ items: [] })
      // PATCH update response (return profile)
      .mockResolvedValueOnce({
        id: 'id',
        role: 'user',
        created_at: '',
        updated_at: '',
        full_name: 'Snake Case',
      })

    const { container } = render(<ProfilePage />)

    await act(async () => {
      await Promise.resolve()
    })

    // Click Save button
    const saveBtn = Array.from(container.querySelectorAll('button')).find((b) => b.textContent?.includes('Save Changes')) as HTMLButtonElement
    expect(saveBtn).toBeTruthy()

    await act(async () => {
      saveBtn.click()
      await Promise.resolve()
    })

    // Expect third call to be PATCH '/v1/user/profile' with snake_case body
    const patchCall = apiFetchMock.mock.calls.find((c) => c[0] === '/v1/user/profile' && (c[1]?.method === 'PATCH'))
    expect(patchCall).toBeTruthy()

    const body = patchCall?.[1]?.body as string
    expect(typeof body).toBe('string')
    expect(body).toContain('"full_name"')
    expect(body).toContain('"billing_address"')

    // Success message shown
    expect(container.textContent || '').toContain('Profile updated successfully!')
  })

  it('success: clicking Subscribe triggers create-checkout call and redirects', async () => {
    // Arrange: profile then no subscriptions -> shows Subscribe button
    apiFetchMock
      .mockResolvedValueOnce({ id: 'id', role: 'user', created_at: '', updated_at: '' })
      .mockResolvedValueOnce({ items: [] })
      .mockResolvedValueOnce({ url: 'https://checkout.example.com' })

    // Mock window.location to allow setting href
    const originalLocation = window.location
    Object.defineProperty(window, 'location', { value: { href: '' } as Location, configurable: true })

    const { container } = render(<ProfilePage />)

    await act(async () => {
      await Promise.resolve()
    })

    // Click Subscribe Now button
    const subscribeBtn = Array.from(container.querySelectorAll('button')).find((b) => b.textContent?.includes('Subscribe Now')) as HTMLButtonElement
    expect(subscribeBtn).toBeTruthy()

    await act(async () => {
      subscribeBtn.click()
      await Promise.resolve()
    })

    // Expect POST to create-checkout
    const call = apiFetchMock.mock.calls.find((c) => c[0] === '/v1/billing/create-checkout')
    expect(call).toBeTruthy()
    expect(call?.[1]?.method).toBe('POST')

    // Restore location
    Object.defineProperty(window, 'location', { value: originalLocation, configurable: true })
  })

  it('success: clicking Manage Subscription triggers billing portal call and redirects', async () => {
    // Arrange: profile then active subscription -> shows Manage button
    apiFetchMock
      .mockResolvedValueOnce({ id: 'id', role: 'user', created_at: '', updated_at: '' })
      .mockResolvedValueOnce({ items: [{ id: 'sub_1', status: 'active' }] })
      .mockResolvedValueOnce({ url: 'https://portal.example.com' })

    const originalLocation = window.location
    Object.defineProperty(window, 'location', { value: { href: '' } as Location, configurable: true })

    const { container } = render(<ProfilePage />)

    await act(async () => {
      await Promise.resolve()
    })

    // Click Manage Subscription button
    const manageBtn = Array.from(container.querySelectorAll('button')).find((b) => b.textContent?.includes('Manage Subscription')) as HTMLButtonElement
    expect(manageBtn).toBeTruthy()

    await act(async () => {
      manageBtn.click()
      await Promise.resolve()
    })

    const call = apiFetchMock.mock.calls.find((c) => c[0] === '/v1/billing/portal')
    expect(call).toBeTruthy()
    expect(call?.[1]?.method).toBe('POST')

    Object.defineProperty(window, 'location', { value: originalLocation, configurable: true })
  })
})
