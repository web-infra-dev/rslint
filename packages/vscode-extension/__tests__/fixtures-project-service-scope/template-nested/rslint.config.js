// Nested rslint config inside a template-style subdirectory that does NOT
// ship a tsconfig.json. Mirrors the create-rstack layout where the
// `template-rslint/` starter directory carries a rslint.config.ts but no
// tsconfig. Previously this caused the LSP to fall back to a global
// "allow-all" mode, incorrectly enabling type-aware rules on files under
// OTHER configs (e.g. root-level test files outside the root tsconfig's
// include). The fix scopes tsConfigPaths per-config so this only disables
// filtering for files under THIS config's directory.
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
    },
  },
];
