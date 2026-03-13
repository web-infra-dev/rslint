import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint:recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  files: ['**/*.js', '**/*.jsx', '**/*.mjs', '**/*.cjs'],
  rules: {
    'constructor-super': 'error',
    // 'no-control-regex': 'error', // not implemented
    // 'no-delete-var': 'error', // not implemented
    // 'no-dupe-class-members': 'error', // not implemented
    // 'no-dupe-else-if': 'error', // not implemented
    // 'no-empty-character-class': 'error', // not implemented
    // 'no-empty-static-block': 'error', // not implemented
    // 'no-ex-assign': 'error', // not implemented
    // 'no-extra-boolean-cast': 'error', // not implemented
    // 'no-fallthrough': 'error', // not implemented
    // 'no-func-assign': 'error', // not implemented
    // 'no-global-assign': 'error', // not implemented
    // 'no-import-assign': 'error', // not implemented
    // 'no-invalid-regexp': 'error', // not implemented
    // 'no-irregular-whitespace': 'error', // not implemented
    // 'no-misleading-character-class': 'error', // not implemented
    // 'no-new-native-nonconstructor': 'error', // not implemented
    // 'no-nonoctal-decimal-escape': 'error', // not implemented
    // 'no-obj-calls': 'error', // not implemented
    // 'no-octal': 'error', // not implemented
    // 'no-prototype-builtins': 'error', // not implemented
    // 'no-redeclare': 'error', // not implemented
    // 'no-regex-spaces': 'error', // not implemented
    // 'no-self-assign': 'error', // not implemented
    // 'no-setter-return': 'error', // not implemented
    // 'no-shadow-restricted-names': 'error', // not implemented
    // 'no-this-before-super': 'error', // not implemented
    // 'no-undef': 'error', // not implemented
    // 'no-unexpected-multiline': 'error', // not implemented
    // 'no-unreachable': 'error', // not implemented
    // 'no-unsafe-finally': 'error', // not implemented
    // 'no-unsafe-negation': 'error', // not implemented
    // 'no-unsafe-optional-chaining': 'error', // not implemented
    // 'no-unused-labels': 'error', // not implemented
    // 'no-unused-private-class-members': 'error', // not implemented
    // 'no-unused-vars': 'error', // not implemented
    // 'no-useless-assignment': 'error', // not implemented
    // 'no-useless-backreference': 'error', // not implemented
    // 'no-useless-catch': 'error', // not implemented
    // 'no-useless-escape': 'error', // not implemented
    // 'no-with': 'error', // not implemented
    // 'require-yield': 'error', // not implemented
    // 'use-isnan': 'error', // not implemented
    // 'valid-typeof': 'error', // not implemented
    'for-direction': 'error',
    'getter-return': 'error',
    'no-async-promise-executor': 'error',
    'no-case-declarations': 'error',
    'no-class-assign': 'error',
    'no-compare-neg-zero': 'error',
    'no-cond-assign': 'error',
    'no-const-assign': 'error',
    'no-constant-binary-expression': 'error',
    'no-constant-condition': 'error',
    'no-debugger': 'error',
    'no-dupe-args': 'error',
    'no-dupe-keys': 'error',
    'no-duplicate-case': 'error',
    'no-empty': 'error',
    'no-empty-pattern': 'error',
    'no-loss-of-precision': 'error',
    'no-sparse-arrays': 'error',
  },
};

export { recommended };
