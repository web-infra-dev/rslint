import stylistic from '@stylistic/eslint-plugin';

import {
  createRuleTester,
  type RuleTesterPlugin,
} from '../src/util/eslint-plugin-rule-tester';

// @stylistic/eslint-plugin alignment RuleTester — the shared CLI-driven tester
// bound to this plugin. See `../src/util/eslint-plugin-rule-tester.ts`.
export const RuleTester = createRuleTester({
  pkg: '@stylistic/eslint-plugin',
  prefix: '@stylistic',
  plugin: stylistic as unknown as RuleTesterPlugin,
});

export type {
  ValidTestCase,
  InvalidTestCase,
} from '../src/util/eslint-plugin-rule-tester';
