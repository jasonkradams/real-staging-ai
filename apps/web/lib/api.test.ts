import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { apiFetch } from './api';

// Mock global fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('apiFetch', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset to browser environment
    Object.defineProperty(window, 'window', {
      writable: true,
      value: global,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('success: token injection', () => {
    it('success: includes Authorization header when token is available', async () => {
      // Arrange
      const mockToken = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.test-token';
      mockFetch
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ token: mockToken }),
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ data: 'test' }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      await apiFetch('/v1/projects');

      // Assert
      expect(mockFetch).toHaveBeenCalledTimes(2);
      
      // First call to /auth/access-token
      expect(mockFetch).toHaveBeenNthCalledWith(1, '/auth/access-token');
      
      // Second call with Authorization header
      const secondCall = mockFetch.mock.calls[1];
      expect(secondCall[0]).toBe('/api/v1/projects');
      expect(secondCall[1]?.headers).toMatchObject({
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${mockToken}`,
      });
    });

    it('success: uses accessToken field from response', async () => {
      // Arrange
      const mockToken = 'test-access-token';
      mockFetch
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ accessToken: mockToken }),
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ success: true }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      await apiFetch('/v1/test');

      // Assert
      const authHeader = (mockFetch.mock.calls[1][1] as RequestInit)?.headers as Record<string, string>;
      expect(authHeader['Authorization']).toBe(`Bearer ${mockToken}`);
    });

    it('success: uses access_token field from response (snake_case)', async () => {
      // Arrange
      const mockToken = 'snake-case-token';
      mockFetch
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ access_token: mockToken }),
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ success: true }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      await apiFetch('/v1/test');

      // Assert
      const authHeader = (mockFetch.mock.calls[1][1] as RequestInit)?.headers as Record<string, string>;
      expect(authHeader['Authorization']).toBe(`Bearer ${mockToken}`);
    });
  });

  describe('success: request handling', () => {
    it('success: makes API request without token when token fetch fails', async () => {
      // Arrange
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
          status: 401,
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ data: 'test' }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      const result = await apiFetch('/v1/projects');

      // Assert
      expect(result).toEqual({ data: 'test' });
      expect(mockFetch).toHaveBeenCalledTimes(2);
      
      // Second call should not have Authorization header
      const secondCall = mockFetch.mock.calls[1];
      const headers = secondCall[1]?.headers as Record<string, string>;
      expect(headers['Authorization']).toBeUndefined();
    });

    it('success: constructs correct URL with API_BASE', async () => {
      // Arrange
      mockFetch
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ token: 'test-token' }),
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ success: true }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      await apiFetch('/v1/projects');

      // Assert
      const apiCall = mockFetch.mock.calls[1];
      expect(apiCall[0]).toBe('/api/v1/projects');
    });

    it('success: passes custom headers', async () => {
      // Arrange
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ success: true }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      await apiFetch('/v1/test', {
        headers: {
          'X-Custom-Header': 'custom-value',
        },
      });

      // Assert
      const headers = (mockFetch.mock.calls[1][1] as RequestInit)?.headers as Record<string, string>;
      expect(headers['X-Custom-Header']).toBe('custom-value');
      expect(headers['Content-Type']).toBe('application/json');
    });

    it('success: passes request options (method, body, etc.)', async () => {
      // Arrange
      const requestBody = { name: 'Test Project' };
      mockFetch
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ token: 'test-token' }),
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ id: '123' }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      await apiFetch('/v1/projects', {
        method: 'POST',
        body: JSON.stringify(requestBody),
      });

      // Assert
      const apiCall = mockFetch.mock.calls[1];
      expect(apiCall[1]?.method).toBe('POST');
      expect(apiCall[1]?.body).toBe(JSON.stringify(requestBody));
    });
  });

  describe('success: response parsing', () => {
    it('success: returns JSON response when content-type is application/json', async () => {
      // Arrange
      const mockData = { id: '123', name: 'Test' };
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => mockData,
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      const result = await apiFetch('/v1/test');

      // Assert
      expect(result).toEqual(mockData);
    });

    it('success: returns text response when content-type is not JSON', async () => {
      // Arrange
      const mockText = 'Plain text response';
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          text: async () => mockText,
          headers: new Headers({ 'content-type': 'text/plain' }),
        } as Response);

      // Act
      const result = await apiFetch('/v1/test');

      // Assert
      expect(result).toBe(mockText);
    });

    it('success: handles empty response body', async () => {
      // Arrange
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          text: async () => '',
          headers: new Headers({ 'content-type': 'text/plain' }),
        } as Response);

      // Act
      const result = await apiFetch('/v1/test');

      // Assert
      expect(result).toBe('');
    });
  });

  describe('fail: error handling', () => {
    it('fail: throws error when response is not ok', async () => {
      // Arrange
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
        } as Response)
        .mockResolvedValueOnce({
          ok: false,
          status: 404,
          text: async () => 'Not found',
        } as Response);

      // Act & Assert
      await expect(apiFetch('/v1/nonexistent')).rejects.toThrow('Request failed 404: Not found');
    });

    it('fail: throws error with status code and message', async () => {
      // Arrange
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
        } as Response)
        .mockResolvedValueOnce({
          ok: false,
          status: 500,
          text: async () => 'Internal Server Error',
        } as Response);

      // Act & Assert
      await expect(apiFetch('/v1/test')).rejects.toThrow('Request failed 500: Internal Server Error');
    });

    it('fail: handles error when response.text() fails', async () => {
      // Arrange
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
        } as Response)
        .mockResolvedValueOnce({
          ok: false,
          status: 403,
          text: async () => {
            throw new Error('Failed to read response');
          },
        } as unknown as Response);

      // Act & Assert
      await expect(apiFetch('/v1/test')).rejects.toThrow('Request failed 403: ');
    });

    it('fail: throws error on 401 Unauthorized', async () => {
      // Arrange
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
        } as Response)
        .mockResolvedValueOnce({
          ok: false,
          status: 401,
          text: async () => 'Unauthorized',
        } as Response);

      // Act & Assert
      await expect(apiFetch('/v1/protected')).rejects.toThrow('Request failed 401: Unauthorized');
    });
  });

  describe('edge cases', () => {
    it('success: handles server-side rendering (no window)', async () => {
      // Arrange
      const originalWindow = global.window;
      // @ts-expect-error - Testing SSR behavior by deleting window
      delete global.window;

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: 'test' }),
        headers: new Headers({ 'content-type': 'application/json' }),
      } as Response);

      // Act
      const result = await apiFetch('/v1/test');

      // Assert
      expect(result).toEqual({ data: 'test' });
      expect(mockFetch).toHaveBeenCalledTimes(1); // No token fetch attempt

      // Restore
      global.window = originalWindow;
    });

    it('success: handles token fetch error gracefully', async () => {
      // Arrange
      mockFetch
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ data: 'test' }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      const result = await apiFetch('/v1/test');

      // Assert
      expect(result).toEqual({ data: 'test' });
      expect(mockFetch).toHaveBeenCalledTimes(2);
    });

    it('success: handles malformed JSON in token response', async () => {
      // Arrange
      mockFetch
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ notAToken: 'value' }),
        } as Response)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ data: 'test' }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response);

      // Act
      const result = await apiFetch('/v1/test');

      // Assert
      expect(result).toEqual({ data: 'test' });
      // Should proceed without token
      const headers = (mockFetch.mock.calls[1][1] as RequestInit)?.headers as Record<string, string>;
      expect(headers['Authorization']).toBeUndefined();
    });
  });
});
