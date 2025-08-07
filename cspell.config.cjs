const { banWords } = require('cspell-ban-words');

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
  ],
  flagWords: banWords,
  dictionaries: ['dictionary'],
  dictionaryDefinitions: [
    {
      name: 'dictionary',
      path: './scripts/dictionary.txt',
      addWords: true,
    },
  ],
};
