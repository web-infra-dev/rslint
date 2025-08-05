import { defineConfig } from '@rstest/core';

export default defineConfig({
  testEnvironment: 'node',
  globals: true,
  include: [
    './tests/typescript-eslint/rules/await-thenable.test.ts',
    './tests/typescript-eslint/rules/no-array-delete.test.ts',
    './tests/typescript-eslint/rules/no-for-in-array.test.ts',
  ],
});
