// Matches the shape of `ts.configs.recommended`: parserOptions only sets
// `projectService: true`, with no explicit `project`. tsconfig.json's
// `include` covers `src` only.
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
