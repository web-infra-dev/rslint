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
]);

// rslint validates every config entry's rule options at load time
// regardless of `files`, so an invalid entry can't live in the array above
// without breaking extension activation. rslint's config loader only reads
// the module's default export, so this named export is never loaded — it
// exists purely for TypeScript to type-check.
export const typeCheckOnly = defineConfig([
  {
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
