import { define } from 'rstack';

define.lint([
  {
    languageOptions: { parserOptions: { projectService: true } },
    rules: {
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/no-unsafe-assignment': 'error',
      'no-debugger': 'error',
      'no-var': 'error',
      'prefer-const': 'error',
    },
  },
]);
