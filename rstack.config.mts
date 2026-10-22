import { define } from 'rstack';

define.fmt({
  singleQuote: true,
  ignorePatterns: [
    '.vscode/',
    '.agents/skills/create-draft-release-notes',
    // Immutable compiler resources and frozen oracle inputs retain their source bytes.
    'vendor/typescript-libs/',
    'scripts/tsrs/oracle-cases/',
    'packages/rslint-api/src/ast-compat/',
    'packages/rslint-test-tools/tests/typescript-eslint/fixtures',
    'packages/rslint-test-tools/tests/typescript-eslint/rules',
    'binaries/',
    'website/releases.json',
    'packages/vscode-extension/__tests__/fixtures-monorepo/packages/broken/',
  ],
});

define.staged({
  '*.{js,jsx,ts,tsx,mjs,cjs,mts,cts}': ['rs fmt'],
  '*.{json,jsonc,md,mdx,css,scss,less,html,yml,yaml}': ['rs fmt'],
});
