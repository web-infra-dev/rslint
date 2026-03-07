import type { RslintConfigEntry } from '../define-config.js';

// Preset for JavaScript projects with all implemented core ESLint rules.
const recommended: RslintConfigEntry = {
  files: ['**/*.js', '**/*.jsx', '**/*.mjs', '**/*.cjs'],
  rules: {
    'array-callback-return': 'error',
    'constructor-super': 'error',
    'default-case': 'error',
    'for-direction': 'error',
    'getter-return': 'error',
    'no-async-promise-executor': 'error',
    'no-await-in-loop': 'error',
    'no-case-declarations': 'error',
    'no-class-assign': 'error',
    'no-compare-neg-zero': 'error',
    'no-cond-assign': 'error',
    'no-console': 'error',
    'no-const-assign': 'error',
    'no-constant-binary-expression': 'error',
    'no-constant-condition': 'error',
    'no-constructor-return': 'error',
    'no-debugger': 'error',
    'no-dupe-args': 'error',
    'no-dupe-keys': 'error',
    'no-duplicate-case': 'error',
    'no-empty': 'error',
    'no-empty-pattern': 'error',
    'no-loss-of-precision': 'error',
    'no-template-curly-in-string': 'error',
    'no-sparse-arrays': 'error',
  },
};

export const js = { configs: { recommended } };
