import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint:recommended (@eslint/js@10.x).
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  files: ['**/*.js', '**/*.jsx', '**/*.mjs', '**/*.cjs'],
  rules: {
    'constructor-super': 'error',
    'no-control-regex': 'error',
    'no-delete-var': 'error',
    'no-dupe-class-members': 'error',
    'no-dupe-else-if': 'error',
    'no-empty-character-class': 'error',
    // 'no-empty-static-block': 'error', // not implemented
    'no-ex-assign': 'error',
    'no-extra-boolean-cast': 'error',
    'no-fallthrough': 'error',
    'no-func-assign': 'error',
    'no-global-assign': 'error',
    'no-import-assign': 'error',
    'no-invalid-regexp': 'error',
    // 'no-irregular-whitespace': 'error', // not implemented
    'no-misleading-character-class': 'error',
    // 'no-new-native-nonconstructor': 'error', // not implemented
    // 'no-nonoctal-decimal-escape': 'error', // not implemented
    'no-obj-calls': 'error',
    'no-octal': 'error',
    'no-prototype-builtins': 'error',
    // 'no-redeclare': 'error', // not implemented
    'no-regex-spaces': 'error',
    'no-self-assign': 'error',
    'no-setter-return': 'error',
    'no-shadow-restricted-names': 'error',
    'no-this-before-super': 'error',
    'no-undef': 'error',
    // 'no-unexpected-multiline': 'error', // not implemented
    'no-unreachable': 'error',
    'no-unsafe-finally': 'error',
    'no-unsafe-negation': 'error',
    'no-unsafe-optional-chaining': 'error',
    // 'no-unused-labels': 'error', // not implemented
    // 'no-unused-private-class-members': 'error', // not implemented
    // 'no-unused-vars': 'error', // not implemented
    // 'no-unassigned-vars': 'error', // not implemented
    // 'no-useless-assignment': 'error', // not implemented
    // 'no-useless-backreference': 'error', // not implemented
    'no-useless-catch': 'error',
    // 'no-useless-escape': 'error', // not implemented
    'no-with': 'error',
    // 'preserve-caught-error': 'error', // not implemented
    'require-yield': 'error',
    'use-isnan': 'error',
    'valid-typeof': 'error',
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
