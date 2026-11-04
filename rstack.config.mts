import { define } from 'rstack';

define.fmt({
  singleQuote: true,
  ignorePatterns: [
    '.vscode/',
    '.agents/skills/create-draft-release-notes',
    'typescript-go/',
    'packages/rslint-test-tools/tests/typescript-eslint/fixtures',
    'packages/rslint-test-tools/tests/typescript-eslint/rules',
    'binaries/',
    'internal/**/rules/**/*.md',
    'website/releases.json',
  ],
});

define.staged({
  '*.{js,jsx,ts,tsx,mjs,cjs,mts,cts}': ['rs fmt'],
  '*.{json,jsonc,md,mdx,css,scss,less,html,yml,yaml}': ['rs fmt'],
});
