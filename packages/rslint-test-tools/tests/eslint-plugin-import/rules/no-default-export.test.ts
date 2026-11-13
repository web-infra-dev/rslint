// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-default-export', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'module.exports = function foo() {}',
    },
    {
      name: 'run 0 valid 1',
      code: 'module.exports = function foo() {}',
    },
    {
      name: 'run 0 valid 2',
      code: "\n        export const foo = 'foo';\n        export const bar = 'bar';\n      ",
    },
    {
      name: 'run 0 valid 3',
      code: "\n        export const foo = 'foo';\n        export function bar() {};\n      ",
    },
    {
      name: 'run 0 valid 4',
      code: "export const foo = 'foo';",
    },
    {
      name: 'run 0 valid 5',
      code: "\n        const foo = 'foo';\n        export { foo };\n      ",
    },
    {
      name: 'run 0 valid 6',
      code: 'let foo, bar; export { foo, bar }',
    },
    {
      name: 'run 0 valid 7',
      code: 'export const { foo, bar } = item;',
    },
    {
      name: 'run 0 valid 8',
      code: 'export const { foo, bar: baz } = item;',
    },
    {
      name: 'run 0 valid 9',
      code: 'export const { foo: { bar, baz } } = item;',
    },
    {
      name: 'run 0 valid 10',
      code: '\n        let item;\n        export const foo = item;\n        export { item };\n      ',
    },
    {
      name: 'run 0 valid 11',
      code: "export * from './foo';",
    },
    {
      name: 'run 0 valid 12',
      code: 'export const { foo } = { foo: "bar" };',
    },
    {
      name: 'run 0 valid 13',
      code: 'export const { foo: { bar } } = { foo: { bar: "baz" } };',
    },
    {
      name: 'run 0 valid 14',
      code: 'export { a, b } from "foo.js"',
    },
    {
      name: 'run 0 valid 15',
      code: "import * as foo from './foo';",
    },
    {
      name: 'run 0 valid 16',
      code: "import foo from './foo';",
    },
    {
      name: 'run 0 valid 17',
      code: "import {default as foo} from './foo';",
    },
    {
      name: 'run 0 valid 18',
      code: 'export type UserId = number;',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'export default function bar() {};',
      errors: [
        {
          messageId: 'preferNamed',
          message: 'Prefer named exports.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "\n        export const foo = 'foo';\n        export default bar;",
      errors: [
        {
          messageId: 'preferNamed',
          message: 'Prefer named exports.',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'export default class Bar {};',
      errors: [
        {
          messageId: 'preferNamed',
          message: 'Prefer named exports.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'export default function() {};',
      errors: [
        {
          messageId: 'preferNamed',
          message: 'Prefer named exports.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'export default class {};',
      errors: [
        {
          messageId: 'preferNamed',
          message: 'Prefer named exports.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'let foo; export { foo as default }',
      errors: [
        {
          messageId: 'noAliasDefault',
          message:
            'Do not alias `foo` as `default`. Just export `foo` itself instead.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'let foo; export { foo as "default" }',
      errors: [
        {
          messageId: 'noAliasDefault',
          message:
            'Do not alias `foo` as `default`. Just export `foo` itself instead.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
  ],
});
