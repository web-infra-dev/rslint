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
    },
  },
];
