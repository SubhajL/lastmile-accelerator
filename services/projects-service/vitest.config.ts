import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    passWithNoTests: false,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: ['node_modules/', 'dist/'],
      lines: 80,
      functions: 80,
      branches: 75,
      statements: 80,
    },
  },
});
