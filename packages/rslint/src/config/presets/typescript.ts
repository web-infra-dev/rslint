import type { RslintConfigEntry } from '../define-config.js';

const base: RslintConfigEntry = {
  files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'],
  plugins: ['@typescript-eslint'],
};

// Aligned with official @typescript-eslint/recommended.
// Includes the eslint-recommended override layer (disables core rules handled by TS,
// enables TS-beneficial rules).
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  ...base,
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
    'no-dupe-class-members': 'off',
    'no-dupe-keys': 'off',
    'no-func-assign': 'off',
    'no-import-assign': 'off',
    'no-new-native-nonconstructor': 'off',
    'no-new-symbol': 'off',
    'no-obj-calls': 'off',
    'no-redeclare': 'off',
    'no-setter-return': 'off',
    'no-this-before-super': 'off',
    'no-undef': 'off',
    'no-unreachable': 'off',
    'no-unsafe-negation': 'off',
    'no-with': 'off',

    // TS-beneficial rules enabled by eslint-recommended override
    'no-var': 'error',
    'prefer-const': 'error',
    'prefer-rest-params': 'error',
    'prefer-spread': 'error',

    // Remaining eslint:recommended rules (not turned off by TS)
    'no-control-regex': 'error',
    'no-delete-var': 'error',
    'no-dupe-else-if': 'error',
    'no-empty-character-class': 'error',
    'no-empty-static-block': 'error',
    'no-ex-assign': 'error',
    'no-extra-boolean-cast': 'error',
    'no-fallthrough': 'error',
    'no-global-assign': 'error',
    'no-invalid-regexp': 'error',
    'no-irregular-whitespace': 'error',
    'no-misleading-character-class': 'error',
    'no-nonoctal-decimal-escape': 'error',
    'no-octal': 'error',
    'no-prototype-builtins': 'error',
    'no-regex-spaces': 'error',
    'no-self-assign': 'error',
    'no-shadow-restricted-names': 'error',
    'no-unexpected-multiline': 'error',
    'no-unsafe-finally': 'error',
    'no-unsafe-optional-chaining': 'error',
    'no-unused-labels': 'error',
    'no-unused-private-class-members': 'error',
    'no-unassigned-vars': 'error',
    // 'no-useless-assignment': 'error', // not implemented
    'no-useless-backreference': 'error',
    'no-useless-catch': 'error',
    'no-useless-escape': 'error',
    // 'preserve-caught-error': 'error', // not implemented
    'require-yield': 'error',
    'use-isnan': 'error',
    'valid-typeof': 'error',
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
    '@typescript-eslint/no-empty-object-type': 'error',
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/no-extra-non-null-assertion': 'error',
    '@typescript-eslint/no-misused-new': 'error',
    '@typescript-eslint/no-namespace': 'error',
    '@typescript-eslint/no-non-null-asserted-optional-chain': 'error',
    '@typescript-eslint/no-require-imports': 'error',
    '@typescript-eslint/no-this-alias': 'error',
    '@typescript-eslint/no-unnecessary-type-constraint': 'error',
    '@typescript-eslint/no-unsafe-declaration-merging': 'error',
    '@typescript-eslint/no-unsafe-function-type': 'error',
    'no-unused-expressions': 'off',
    '@typescript-eslint/no-unused-expressions': 'error',
    'no-unused-vars': 'off',
    // Differs from typescript-eslint recommended (which uses bare 'error').
    // Ignoring _-prefixed vars/args is a widely adopted community convention,
    // so we include it in our default recommended config for better DX.
    '@typescript-eslint/no-unused-vars': [
      'error',
      { varsIgnorePattern: '^_', argsIgnorePattern: '^_' },
    ],
    '@typescript-eslint/no-wrapper-object-types': 'error',
    '@typescript-eslint/prefer-as-const': 'error',
    '@typescript-eslint/prefer-namespace-keyword': 'error',
    '@typescript-eslint/triple-slash-reference': 'error',
  },
};

export { base, recommended };
