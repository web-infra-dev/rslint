import { defineConfig } from '@rstest/core';

export default defineConfig({
  testEnvironment: 'node',
  globals: true,
  include: [
    './tests/typescript-eslint/rules/adjacent-overload-signatures.test.ts',
    './tests/typescript-eslint/rules/array-type.test.ts',
    './tests/typescript-eslint/rules/await-thenable.test.ts',
    './tests/typescript-eslint/rules/class-literal-property-style.test.ts',
    './tests/typescript-eslint/rules/no-array-delete.test.ts',
    './tests/typescript-eslint/rules/no-confusing-void-expression.test.ts',
    './tests/typescript-eslint/rules/no-empty-function.test.ts',
    './tests/typescript-eslint/rules/no-empty-interface.test.ts',
    './tests/typescript-eslint/rules/no-require-imports.test.ts',
  ],
});
