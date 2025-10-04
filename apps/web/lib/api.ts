export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || '/api'

/**
 * Get the access token from the Auth0 session
 * This fetches the token from the /auth/access-token endpoint
 */
async function getAccessToken(): Promise<string | null> {
  if (typeof window === 'undefined') return null
  
  try {
    const response = await fetch('/auth/access-token');
    if (!response.ok) return null;
    
    const data = await response.json();
    // Auth0 SDK returns the token under the 'token' property
    return data.token || data.accessToken || data.access_token || null;
  } catch (error) {
    console.error('Failed to get access token:', error);
    return null;
  }
}

type Json = string | number | boolean | null | Json[] | { [key: string]: Json };

/**
 * API client for making authenticated requests to the backend API
 * Automatically includes the Auth0 access token in the Authorization header
 */
export async function apiFetch<T = Json>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE}${path}`
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  }
  
  // Get the access token from Auth0 session
  const token = await getAccessToken()
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
