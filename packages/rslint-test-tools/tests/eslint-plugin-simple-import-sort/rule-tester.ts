import simpleImportSort from 'eslint-plugin-simple-import-sort';

import {
  createRuleTester,
  type RuleTesterPlugin,
} from '../src/util/eslint-plugin-rule-tester';

// eslint-plugin-simple-import-sort alignment RuleTester — the shared CLI-driven
// tester bound to this plugin. See `../src/util/eslint-plugin-rule-tester.ts`.
export const RuleTester = createRuleTester({
  pkg: 'eslint-plugin-simple-import-sort',
  prefix: 'simple-import-sort',
  plugin: simpleImportSort as unknown as RuleTesterPlugin,
});

export type {
  ValidTestCase,
  InvalidTestCase,
} from '../src/util/eslint-plugin-rule-tester';
