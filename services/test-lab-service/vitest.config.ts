import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    clearMocks: true,
    restoreMocks: true,
    mockReset: true,
    // Ensure ESM modules are handled correctly
    setupFiles: [],
  },
});
