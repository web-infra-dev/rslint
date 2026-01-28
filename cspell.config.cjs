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
    'packages/rslint/pkg/mod',
    'cmd/tsgo',
    './agents',
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
