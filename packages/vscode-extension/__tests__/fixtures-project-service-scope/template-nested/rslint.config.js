// Nested rslint config inside a template-style subdirectory that does NOT
// ship a tsconfig.json. Mirrors the create-rstack layout where the
// `template-rslint/` starter directory carries a rslint.config.ts but no
// tsconfig. Previously this caused the LSP to fall back to a global
// "allow-all" mode, incorrectly enabling type-aware rules. A config without
// any resolved tsconfig now disables type-aware rules for its own files.
export default [
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
      'no-var': 'error',
    },
  },
];
