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
      '@typescript-eslint/no-unused-vars': 'warn',

      // --- Additional rules not in recommended ---
      '@typescript-eslint/non-nullable-type-assertion-style': 'warn',
      '@typescript-eslint/promise-function-async': 'warn',
      '@typescript-eslint/no-floating-promises': 'warn',
      '@typescript-eslint/no-unsafe-return': 'warn',
      '@typescript-eslint/return-await': 'warn',
      '@typescript-eslint/no-unsafe-member-access': 'warn',
      '@typescript-eslint/no-unsafe-argument': 'warn',
      '@typescript-eslint/no-unsafe-assignment': 'warn',
      '@typescript-eslint/no-unsafe-call': 'warn',
      '@typescript-eslint/no-confusing-void-expression': 'warn',
      '@typescript-eslint/no-redundant-type-constituents': 'warn',
      '@typescript-eslint/no-unsafe-type-assertion': 'warn',
      '@typescript-eslint/no-unsafe-enum-comparison': 'warn',
      '@typescript-eslint/no-unnecessary-type-assertion': 'warn',
      '@typescript-eslint/no-var-requires': 'error',
      '@typescript-eslint/no-empty-function': 'error',
      '@typescript-eslint/no-empty-interface': 'error',
      '@typescript-eslint/no-misused-promises': 'warn',
      '@typescript-eslint/require-await': 'warn',
      '@typescript-eslint/prefer-readonly': 'warn',
      '@typescript-eslint/no-non-null-assertion': 'warn',
      'prefer-const': 'warn',
      '@typescript-eslint/no-dynamic-delete': 'off',
      '@typescript-eslint/prefer-includes': 'off',
      '@typescript-eslint/prefer-regexp-exec': 'off',
      '@typescript-eslint/prefer-ts-expect-error': 'off',
      'no-console': 'warn',
    },
  },
]);
