import { defineConfig } from '@rstest/core';

export default defineConfig({
  testEnvironment: 'node',
  globals: true,
  include: [
    // cli
    './tests/cli/basic.test.ts',

    // eslint-plugin-import
    './tests/eslint-plugin-import/rules/no-self-import.test.ts',
    './tests/eslint-plugin-import/rules/no-webpack-loader-syntax.test.ts',

    // typescript-eslint
    './tests/typescript-eslint/rules/adjacent-overload-signatures.test.ts',
    './tests/typescript-eslint/rules/array-type.test.ts',
    './tests/typescript-eslint/rules/await-thenable.test.ts',
    './tests/typescript-eslint/rules/class-literal-property-style.test.ts',
    './tests/typescript-eslint/rules/dot-notation.test.ts',
    './tests/typescript-eslint/rules/no-array-delete.test.ts',
    // too many autofix errors
    // './tests/typescript-eslint/rules/no-confusing-void-expression.test.ts',
    './tests/typescript-eslint/rules/no-empty-function.test.ts',
    './tests/typescript-eslint/rules/no-empty-interface.test.ts',
    './tests/typescript-eslint/rules/no-explicit-any.test.ts',
    './tests/typescript-eslint/rules/no-floating-promises.test.ts',
    './tests/typescript-eslint/rules/no-require-imports.test.ts',
    // too many autofix errors
    './tests/typescript-eslint/rules/no-duplicate-type-constituents.test.ts',
    './tests/typescript-eslint/rules/no_namespace.test.ts',
    './tests/typescript-eslint/rules/no-implied-eval.test.ts',
  ],
});
