import { defineConfig } from '@rstest/core';

export default defineConfig({
  testEnvironment: 'node',
  globals: true,
  include: [
    // eslint-plugin-import
    './tests/eslint-plugin-import/rules/no-self-import.test.ts',

    // typescript-eslint
    './tests/typescript-eslint/rules/adjacent-overload-signatures.test.ts',
    './tests/typescript-eslint/rules/array-type.test.ts',
    './tests/typescript-eslint/rules/await-thenable.test.ts',
    './tests/typescript-eslint/rules/class-literal-property-style.test.ts',
    './tests/typescript-eslint/rules/dot-notation.test.ts',
    './tests/typescript-eslint/rules/explicit-member-accessibility.test.ts',
    './tests/typescript-eslint/rules/max-params.test.ts',
    './tests/typescript-eslint/rules/member-ordering.test.ts',
    './tests/typescript-eslint/rules/member-ordering/member-ordering-alphabetically-case-insensitive-order.test.ts',
    './tests/typescript-eslint/rules/member-ordering/member-ordering-alphabetically-order.test.ts',
    './tests/typescript-eslint/rules/member-ordering/member-ordering-natural-case-insensitive-order.test.ts',
    './tests/typescript-eslint/rules/member-ordering/member-ordering-natural-order.test.ts',
    './tests/typescript-eslint/rules/member-ordering/member-ordering-optionalMembers.test.ts',
    './tests/typescript-eslint/rules/no-array-delete.test.ts',
    './tests/typescript-eslint/rules/no-confusing-void-expression.test.ts',
    './tests/typescript-eslint/rules/no-empty-function.test.ts',
    './tests/typescript-eslint/rules/no-empty-interface.test.ts',
    './tests/typescript-eslint/rules/no-require-imports.test.ts',
    './tests/typescript-eslint/rules/no-duplicate-type-constituents.test.ts',
    './tests/typescript-eslint/rules/no_namespace.test.ts',
    './tests/typescript-eslint/rules/no-implied-eval.test.ts',
  ],
});
