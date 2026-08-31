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
    'website/rule-releases.json',
    'packages/vscode-extension/__tests__/fixtures-monorepo/packages/broken/',
  ],
});
