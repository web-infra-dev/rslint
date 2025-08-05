import { defineConfig } from '@rstest/core';

export default defineConfig({
  testEnvironment: 'node',
  globals: true,
  include: ['./tests/typescript-eslint/rules/await-thenable.test.ts', "./tests/typescript-eslint/rules/array-type.test.ts"],
});
