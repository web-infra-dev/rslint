import { defineConfig, ts } from '@rslint/core';

export default defineConfig([
  {
    ignores: [
      'node_modules/**',
      '**/dist/**',
      'typescript-go/**',

      // Test fixtures — not real source code
      '**/fixtures/**',
      'packages/rslint-test-tools/tests/**',

      // VSCode extension test artifacts
      'packages/vscode-extension/__tests__/**',
      'packages/vscode-extension/.vscode-test/**',

      // Generated / build artifacts
      'website/doc_build/**',
      'website/generated/**',

      // Files that need special handling
      'packages/rslint-wasm/src/worker.ts',
      'packages/rule-tester/src/index.ts',
    ],
  },
  // Start from recommended preset, then override rules and parserOptions
  ts.configs.recommended,
  {
    files: ['**/*.ts', '**/*.tsx'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: [
          './packages/*/tsconfig.build.json',
          './packages/*/tsconfig.spec.json',
          './packages/rslint/fixtures/tsconfig.json',
        ],
      },
    },
    rules: {
      // --- Override recommended rules to warn ---
      '@typescript-eslint/no-explicit-any': 'warn',

      // --- Additional rules not in recommended ---
      '@typescript-eslint/non-nullable-type-assertion-style': 'error',
      '@typescript-eslint/promise-function-async': 'error',
      '@typescript-eslint/no-floating-promises': 'warn',
      '@typescript-eslint/no-unsafe-return': 'warn',
      '@typescript-eslint/return-await': 'error',
      '@typescript-eslint/no-unsafe-member-access': 'warn',
      '@typescript-eslint/no-unsafe-argument': 'warn',
      '@typescript-eslint/no-unsafe-assignment': 'warn',
      '@typescript-eslint/no-unsafe-call': 'error',
      '@typescript-eslint/no-confusing-void-expression': 'error',
      '@typescript-eslint/no-redundant-type-constituents': 'error',
      '@typescript-eslint/no-unsafe-type-assertion': 'error',
      '@typescript-eslint/no-unsafe-enum-comparison': 'error',
      '@typescript-eslint/no-unnecessary-type-assertion': 'error',
      '@typescript-eslint/no-var-requires': 'error',
      '@typescript-eslint/no-empty-function': 'error',
      '@typescript-eslint/no-empty-interface': 'error',
      '@typescript-eslint/no-misused-promises': 'warn',
      '@typescript-eslint/require-await': 'warn',
      '@typescript-eslint/prefer-readonly': 'error',
      '@typescript-eslint/no-non-null-assertion': 'warn',
      '@typescript-eslint/no-dynamic-delete': 'error',
      '@typescript-eslint/prefer-includes': 'error',
      '@typescript-eslint/prefer-regexp-exec': 'error',
      '@typescript-eslint/prefer-ts-expect-error': 'error',
      'no-console': ['error', { allow: ['warn', 'error'] }],
    },
  },
]);
