export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./packages/core/tsconfig.lint.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-unsafe-member-access': 'error',
      '@typescript-eslint/require-await': 'error',
      'no-console': 'error',
    },
    plugins: ['@typescript-eslint'],
  },
];
