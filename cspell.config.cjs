module.exports = {
  version: '0.2',
  language: 'en',
  files: ['**/*.{ts,tsx,js,jsx,mjs,cjs,md,mdx,json,go,sh,yml,yaml}'],
  enableFiletypes: ['mdx'],
  ignorePaths: [
    'dist',
    'dist-*',
    'coverage',
    'doc_build',
    'typescript-go',
    'node_modules',
    'pnpm-lock.yaml',
    'shim',
    'packages/vscode-extension/out',
    'packages/rslint-test-tools/tests',
    'packages/rslint/THIRD-PARTY-NOTICES.md',
    'packages/rslint/pkg/mod',
    'packages/rslint/rule-schemas.json',
    'cmd/tsgo',
    './agents',
    'crates/rslint-native/index.js',
    'crates/rslint-native/index.d.ts',
    'internal/linter/*_test.go',
    'internal/utils/minimatch3/*_test.go',
    'internal/lsp/*_test.go',
    'internal/config/*_test.go',
    'internal/rules/valid_typeof/valid_typeof.md',
    'internal/plugins/jsx_a11y/rules/aria_props/aria_props.md',
    'internal/plugins/jsx_a11y/rules/aria_proptypes/aria_proptypes.md',
    'internal/plugins/react/rules/no_typos/no_typos.md',
    'internal/plugins/react/rules/no_unused_prop_types/no_unused_prop_types_upstream.json',
    'website/docs/en/rules/*/',
  ],
  overrides: [
    {
      filename:
        'internal/plugins/node/rules/no_deprecated_api/no_deprecated_api.schema.json',
      // Fixed upstream API identifier in the option enum.
      // cspell:ignore freelist
      ignoreRegExpList: ['/"freelist"/g'],
    },
  ],
  dictionaries: ['dictionary'],
  dictionaryDefinitions: [
    {
      name: 'dictionary',
      path: './scripts/dictionary.txt',
      addWords: true,
    },
  ],
};
