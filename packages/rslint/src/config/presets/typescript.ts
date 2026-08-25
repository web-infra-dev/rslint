import type { RslintConfigEntry } from '../define-config.js';

const base: RslintConfigEntry = {
  languageOptions: {
    parserOptions: {
      projectService: true,
    },
  },
  plugins: ['@typescript-eslint'],
};

// Aligned with official @typescript-eslint/recommended.
// Includes the eslint-recommended override layer (disables core rules handled by TS,
// enables TS-beneficial rules).
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry[] = [
  base,
  {
    files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'],
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
    },
  },
  {
    rules: {
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
  },
];

export { base, recommended };
