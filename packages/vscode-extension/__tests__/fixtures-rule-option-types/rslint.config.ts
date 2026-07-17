import { defineConfig } from '@rslint/core';

export default defineConfig([
  {
    files: ['src/**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      'no-console': ['error', { allow: ['warn'] }],
    },
  },
  {
    // Matches no real file — only here to type-check `rules` below.
    files: ['__type-check-only__/**'],
    rules: {
      // @ts-expect-error `allow` must be a string[], not a number.
      'no-console': ['error', { allow: 123 }],
    },
  },
]);

// Unsuppressed type error — the e2e test's readiness signal that TypeScript
// finished analyzing this file, since the error above is swallowed by
// `@ts-expect-error` on success. `export`ed so it isn't flagged as unused.
export const tsReadySentinel: string = 123;
