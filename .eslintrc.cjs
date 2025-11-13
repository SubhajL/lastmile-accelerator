module.exports = {
  root: true,
  parser: '@typescript-eslint/parser',
  plugins: ['@typescript-eslint'],
  extends: ['eslint:recommended', 'plugin:@typescript-eslint/recommended', 'prettier'],
  env: { node: true, es2022: true },
  parserOptions: { ecmaVersion: 'latest', sourceType: 'module' },
  ignorePatterns: [
    '**/dist/**',
    '**/.next/**',
    '**/node_modules/**',
    '**/target/**',
    'lma-devstack-compose-gitea4001/**',
  ],
  overrides: [
    {
      files: ['**/*.test.ts', '**/*.spec.ts', '**/*.test.tsx', '**/*.spec.tsx'],
      rules: {
        '@typescript-eslint/no-explicit-any': 'off',
        '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
      },
    },
  ],
};

