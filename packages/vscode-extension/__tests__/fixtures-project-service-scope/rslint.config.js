// Explicitly enable projectService without an explicit project.
// The tsconfig includes src only, so other files exercise gap linting.
export default [
  { ignores: ['**/dist/**'] },
  {
    files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'],
    plugins: ['@typescript-eslint'],
    languageOptions: {
      parserOptions: {
        projectService: true,
      },
    },
    rules: {
      '@typescript-eslint/no-unused-vars': 'error',
      // Non-type-aware marker rule. The suite relies on its diagnostic as a
      // "rslint has finished linting this file" signal, so the negative
      // assertion does not need a fixed-duration sleep.
      'no-console': 'error',
    },
  },
];
