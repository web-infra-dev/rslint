// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('unambiguous', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'function x() {}',
      languageOptions: {
        sourceType: 'script',
      },
    },
    {
      name: 'run 0 valid 1',
      code: '"use strict"; function y() {}',
      languageOptions: {
        sourceType: 'script',
      },
    },
    {
      name: 'run 0 valid 2',
      code: 'import y from "z"; function x() {}',
    },
    {
      name: 'run 0 valid 3',
      code: 'import * as y from "z"; function x() {}',
    },
    {
      name: 'run 0 valid 4',
      code: 'import { y } from "z"; function x() {}',
    },
    {
      name: 'run 0 valid 5',
      code: 'import z, { y } from "z"; function x() {}',
    },
    {
      name: 'run 0 valid 6',
      code: 'function x() {}; export {}',
    },
    {
      name: 'run 0 valid 7',
      code: 'function x() {}; export { x }',
    },
    {
      name: 'run 0 valid 8',
      code: 'function x() {}; export { y } from "z"',
    },
    {
      name: 'run 0 valid 9',
      code: 'function x() {}; export * as y from "z"',
    },
    {
      name: 'run 0 valid 10',
      code: 'export function x() {}',
    },
    {
      name: 'run 0 valid 11',
      code: "import 'foo'\nfunction x() { return 42 }",
    },
    {
      name: 'run 0 valid 12',
      code: 'export function x() { return 42 }',
    },
    {
      name: 'run 0 valid 13',
      code: "(function x() { return 42 })()\nexport {} // simple way to mark side-effects-only file as 'module' without any imports/exports",
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'function x() {}',
      languageOptions: {
        sourceType: 'module',
      },
      errors: [
        {
          messageId: '',
          message: 'This module could be parsed as a valid script.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: '(function x() { return 42 })()',
      errors: [
        {
          messageId: '',
          message: 'This module could be parsed as a valid script.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
  ],
});
