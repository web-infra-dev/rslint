export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./packages/core/tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/require-await': 'error',
      'no-console': 'error',
    },
    plugins: ['@typescript-eslint'],
  },
];
