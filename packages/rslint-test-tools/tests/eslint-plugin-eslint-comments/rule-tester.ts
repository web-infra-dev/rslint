import eslintCommentsModule from '@eslint-community/eslint-plugin-eslint-comments';

import {
  createRuleTester,
  type RuleTesterPlugin,
} from '../src/util/eslint-plugin-rule-tester';

// @eslint-community/eslint-plugin-eslint-comments alignment RuleTester — the
// shared CLI-driven tester bound to this plugin. The plugin is CJS, so unwrap
// the ESM default-interop before reading `rules`. See
// `../src/util/eslint-plugin-rule-tester.ts`.
const plugin = ((eslintCommentsModule as { default?: unknown }).default ??
  eslintCommentsModule) as RuleTesterPlugin;

export const RuleTester = createRuleTester({
  pkg: '@eslint-community/eslint-plugin-eslint-comments',
  prefix: 'eslint-comments',
  plugin,
});

export type {
  ValidTestCase,
  InvalidTestCase,
} from '../src/util/eslint-plugin-rule-tester';
