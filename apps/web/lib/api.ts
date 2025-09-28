export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || '/api'

function getToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem('token')
}

type Json = string | number | boolean | null | Json[] | { [key: string]: Json };

export async function apiFetch<T = Json>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE}${path}`
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  }
  const token = getToken()
  if (token) {
    ;(headers as Record<string, string>)['Authorization'] = `Bearer ${token}`
  }
  const res = await fetch(url, {
    ...options,
    headers,
  })
  if (!res.ok) {
    const text = await res.text().catch(() => '')
    throw new Error(`Request failed ${res.status}: ${text}`)
  }
  const contentType = res.headers.get('content-type') || ''
  if (contentType.includes('application/json')) {
    return (await res.json()) as T
  }
  return (await res.text()) as T
}
