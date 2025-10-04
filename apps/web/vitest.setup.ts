import { beforeAll, afterEach, afterAll, vi } from 'vitest';

// Setup global test environment
beforeAll(() => {
  // Mock environment variables
  process.env.NEXT_PUBLIC_API_BASE = '/api';
});

afterEach(() => {
  // Clear all mocks after each test
  vi.clearAllMocks();
});

afterAll(() => {
  // Cleanup
  vi.restoreAllMocks();
});
