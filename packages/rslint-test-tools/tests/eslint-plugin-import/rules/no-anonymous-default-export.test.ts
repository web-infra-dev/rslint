// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-anonymous-default-export', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'const foo = 123\nexport default foo',
    },
    {
      name: 'run 0 valid 1',
      code: 'export default function foo() {}',
    },
    {
      name: 'run 0 valid 2',
      code: 'export default class MyClass {}',
    },
    {
      name: 'run 0 valid 3',
      code: 'export default []',
      options: [
        {
          allowArray: true,
        },
      ],
    },
    {
      name: 'run 0 valid 4',
      code: 'export default () => {}',
      options: [
        {
          allowArrowFunction: true,
        },
      ],
    },
    {
      name: 'run 0 valid 5',
      code: 'export default class {}',
      options: [
        {
          allowAnonymousClass: true,
        },
      ],
    },
    {
      name: 'run 0 valid 6',
      code: 'export default function() {}',
      options: [
        {
          allowAnonymousFunction: true,
        },
      ],
    },
    {
      name: 'run 0 valid 7',
      code: 'export default 123',
      options: [
        {
          allowLiteral: true,
        },
      ],
    },
    {
      name: 'run 0 valid 8',
      code: "export default 'foo'",
      options: [
        {
          allowLiteral: true,
        },
      ],
    },
    {
      name: 'run 0 valid 9',
      code: 'export default `foo`',
      options: [
        {
          allowLiteral: true,
        },
      ],
    },
    {
      name: 'run 0 valid 10',
      code: 'export default {}',
      options: [
        {
          allowObject: true,
        },
      ],
    },
    {
      name: 'run 0 valid 11',
      code: 'export default foo(bar)',
      options: [
        {
          allowCallExpression: true,
        },
      ],
    },
    {
      name: 'run 0 valid 12',
      code: 'export default new Foo()',
      options: [
        {
          allowNew: true,
        },
      ],
    },
    {
      name: 'run 0 valid 13',
      code: 'export default 123',
      options: [
        {
          allowLiteral: true,
          allowObject: true,
        },
      ],
    },
    {
      name: 'run 0 valid 14',
      code: 'export default {}',
      options: [
        {
          allowLiteral: true,
          allowObject: true,
        },
      ],
    },
    {
      name: 'run 0 valid 15',
      code: "export * from 'foo'",
    },
    {
      name: 'run 0 valid 16',
      code: 'const foo = 123\nexport { foo }',
    },
    {
      name: 'run 0 valid 17',
      code: 'const foo = 123\nexport { foo as default }',
    },
    {
      name: 'run 0 valid 18',
      code: 'const foo = 123\nexport { foo as "default" }',
    },
    {
      name: 'run 0 valid 19',
      code: 'export default foo(bar)',
    },
    {
      name: 'run 0 valid 20',
      code: 'for (let { foo, bar } of baz) {}',
    },
    {
      name: 'run 0 valid 21',
      code: 'for (let [ foo, bar ] of baz) {}',
    },
    {
      name: 'run 0 valid 22',
      code: 'const { x, y } = bar',
    },
    {
      name: 'run 0 valid 23',
      code: 'const { x, y, ...z } = bar',
    },
    {
      name: 'run 0 valid 24',
      code: 'let x; export { x }',
    },
    {
      name: 'run 0 valid 25',
      code: 'let x; export { x as y }',
    },
    {
      name: 'run 0 valid 26',
      code: 'export const x = null',
    },
    {
      name: 'run 0 valid 27',
      code: 'export var x = null',
    },
    {
      name: 'run 0 valid 28',
      code: 'export let x = null',
    },
    {
      name: 'run 0 valid 29',
      code: 'export default x',
    },
    {
      name: 'run 0 valid 30',
      code: 'export default class x {}',
    },
    {
      name: 'run 0 valid 31',
      code: 'import json from "./data.json"',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 32',
      code: 'import foo from "./foobar.json";',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 33',
      code: 'import foo from "./foobar";',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 34',
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 35',
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 36',
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
    },
    {
      name: 'run 0 valid 37',
      code: 'import { foo } from "./ignore.invalid.extension"',
    },
    {
      name: 'run 0 valid 38',
      code: 'export default function () {}',
      options: [
        {
          allowAnonymousFunction: true,
        },
      ],
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'export default []',
      errors: [
        {
          messageId: '',
          message:
            'Assign array to a variable before exporting as module default',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'export default () => {}',
      errors: [
        {
          messageId: '',
          message:
            'Assign arrow function to a variable before exporting as module default',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'export default class {}',
      errors: [
        {
          messageId: '',
          message: 'Unexpected default export of anonymous class',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'export default function() {}',
      errors: [
        {
          messageId: '',
          message: 'Unexpected default export of anonymous function',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'export default 123',
      errors: [
        {
          messageId: '',
          message:
            'Assign literal to a variable before exporting as module default',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: "export default 'foo'",
      errors: [
        {
          messageId: '',
          message:
            'Assign literal to a variable before exporting as module default',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'export default `foo`',
      errors: [
        {
          messageId: '',
          message:
            'Assign literal to a variable before exporting as module default',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'export default {}',
      errors: [
        {
          messageId: '',
          message:
            'Assign object to a variable before exporting as module default',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'export default foo(bar)',
      options: [
        {
          allowCallExpression: false,
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Assign call result to a variable before exporting as module default',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'export default new Foo()',
      errors: [
        {
          messageId: '',
          message:
            'Assign instance to a variable before exporting as module default',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'export default 123',
      options: [
        {
          allowObject: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Assign literal to a variable before exporting as module default',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'export default function () {}',
      errors: [
        {
          messageId: '',
          message: 'Unexpected default export of anonymous function',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
  ],
});
