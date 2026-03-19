import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official @typescript-eslint/recommended.
// Includes the eslint-recommended override layer (disables core rules handled by TS,
// enables TS-beneficial rules).
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'],
  plugins: ['@typescript-eslint'],
  languageOptions: {
    parserOptions: {
      projectService: true,
    },
  },
  rules: {
    // --- Core ESLint rules (eslint:recommended) ---
    // Rules handled by TypeScript are turned off per the official
    // typescript-eslint eslint-recommended override.
    'constructor-super': 'off',
    'getter-return': 'off',
    'no-class-assign': 'off',
    'no-const-assign': 'off',
    'no-dupe-args': 'off',
    // 'no-dupe-class-members': 'off', // not implemented
    'no-dupe-keys': 'off',
    // 'no-func-assign': 'off', // not implemented
    // 'no-import-assign': 'off', // not implemented
    // 'no-new-native-nonconstructor': 'off', // not implemented
    // 'no-new-symbol': 'off', // not implemented (deprecated, use no-new-native-nonconstructor)
    // 'no-obj-calls': 'off', // not implemented
    // 'no-redeclare': 'off', // not implemented
    // 'no-setter-return': 'off', // not implemented
    // 'no-this-before-super': 'off', // not implemented
    // 'no-undef': 'off', // not implemented
    // 'no-unreachable': 'off', // not implemented
    // 'no-unsafe-negation': 'off', // not implemented
    // 'no-with': 'off', // not implemented

    // TS-beneficial rules enabled by eslint-recommended override
    // 'no-var': 'error', // not implemented
    // 'prefer-const': 'error', // not implemented
    // 'prefer-rest-params': 'error', // not implemented
    // 'prefer-spread': 'error', // not implemented

    // Remaining eslint:recommended rules (not turned off by TS)
    // 'no-control-regex': 'error', // not implemented
    // 'no-delete-var': 'error', // not implemented
    // 'no-dupe-else-if': 'error', // not implemented
    // 'no-empty-character-class': 'error', // not implemented
    // 'no-empty-static-block': 'error', // not implemented
    // 'no-ex-assign': 'error', // not implemented
    // 'no-extra-boolean-cast': 'error', // not implemented
    // 'no-fallthrough': 'error', // not implemented
    // 'no-global-assign': 'error', // not implemented
    // 'no-invalid-regexp': 'error', // not implemented
    // 'no-irregular-whitespace': 'error', // not implemented
    // 'no-misleading-character-class': 'error', // not implemented
    // 'no-nonoctal-decimal-escape': 'error', // not implemented
    // 'no-octal': 'error', // not implemented
    // 'no-prototype-builtins': 'error', // not implemented
    // 'no-regex-spaces': 'error', // not implemented
    // 'no-self-assign': 'error', // not implemented
    // 'no-shadow-restricted-names': 'error', // not implemented
    // 'no-unexpected-multiline': 'error', // not implemented
    // 'no-unsafe-finally': 'error', // not implemented
    // 'no-unsafe-optional-chaining': 'error', // not implemented
    // 'no-unused-labels': 'error', // not implemented
    // 'no-unused-private-class-members': 'error', // not implemented
    // 'no-unassigned-vars': 'error', // not implemented
    // 'no-useless-assignment': 'error', // not implemented
    // 'no-useless-backreference': 'error', // not implemented
    // 'no-useless-catch': 'error', // not implemented
    // 'no-useless-escape': 'error', // not implemented
    // 'preserve-caught-error': 'error', // not implemented
    // 'require-yield': 'error', // not implemented
    // 'use-isnan': 'error', // not implemented
    // 'valid-typeof': 'error', // not implemented
    'for-direction': 'error',
    'no-async-promise-executor': 'error',
    'no-case-declarations': 'error',
    'no-compare-neg-zero': 'error',
    'no-cond-assign': 'error',
    'no-constant-binary-expression': 'error',
    'no-constant-condition': 'error',
    'no-debugger': 'error',
    'no-duplicate-case': 'error',
    'no-empty': 'error',
    'no-empty-pattern': 'error',
    'no-loss-of-precision': 'error',
    'no-sparse-arrays': 'error',

    // --- @typescript-eslint/recommended rules ---
    '@typescript-eslint/ban-ts-comment': 'error',
    'no-array-constructor': 'off',
    '@typescript-eslint/no-array-constructor': 'error',
    '@typescript-eslint/no-duplicate-enum-values': 'error',
    // '@typescript-eslint/no-empty-object-type': 'error', // not implemented
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/no-extra-non-null-assertion': 'error',
    '@typescript-eslint/no-misused-new': 'error',
    '@typescript-eslint/no-namespace': 'error',
    '@typescript-eslint/no-non-null-asserted-optional-chain': 'error',
    '@typescript-eslint/no-require-imports': 'error',
    '@typescript-eslint/no-this-alias': 'error',
    // '@typescript-eslint/no-unnecessary-type-constraint': 'error', // not implemented
    // '@typescript-eslint/no-unsafe-declaration-merging': 'error', // not implemented
    // '@typescript-eslint/no-unsafe-function-type': 'error', // not implemented
    // 'no-unused-expressions': 'off', // not implemented
    // '@typescript-eslint/no-unused-expressions': 'error', // not implemented
    'no-unused-vars': 'off',
    '@typescript-eslint/no-unused-vars': 'error',
    // '@typescript-eslint/no-wrapper-object-types': 'error', // not implemented
    '@typescript-eslint/prefer-as-const': 'error',
    '@typescript-eslint/prefer-namespace-keyword': 'error',
    '@typescript-eslint/triple-slash-reference': 'error',
  },
};

export { recommended };
