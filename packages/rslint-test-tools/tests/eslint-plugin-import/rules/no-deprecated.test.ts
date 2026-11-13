// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-deprecated', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: "import { x } from './fake' ",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 1',
      code: "import bar from './bar'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 2',
      code: "import { fine } from './deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 3',
      code: "import { _undocumented } from './deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 4',
      code: "import { fn } from './deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/docstyle': ['tomdoc'],
      },
    },
    {
      name: 'run 0 valid 5',
      code: "import { fine } from './tomdoc-deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/docstyle': ['tomdoc'],
      },
    },
    {
      name: 'run 0 valid 6',
      code: "import { _undocumented } from './tomdoc-deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/docstyle': ['tomdoc'],
      },
    },
    {
      name: 'run 0 valid 7',
      code: "import * as depd from './deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 8',
      code: "import * as depd from './deprecated'; console.log(depd.fine())",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 9',
      code: "import { deepDep } from './deep-deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 10',
      code: "import { deepDep } from './deep-deprecated'; console.log(deepDep.fine())",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 11',
      code: "import { deepDep } from './deep-deprecated'; function x(deepDep) { console.log(deepDep.MY_TERRIBLE_ACTION) }",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 12',
      code: 'for (let { foo, bar } of baz) {}',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 13',
      code: 'for (let [ foo, bar ] of baz) {}',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 14',
      code: 'const { x, y } = bar',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 15',
      code: 'const { x, y, ...z } = bar',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 16',
      code: 'let x; export { x }',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 17',
      code: 'let x; export { x as y }',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 18',
      code: 'export const x = null',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 19',
      code: 'export var x = null',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 20',
      code: 'export let x = null',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 21',
      code: 'export default x',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 22',
      code: 'export default class x {}',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 23',
      code: 'import json from "./data.json"',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 24',
      code: 'import foo from "./foobar.json";',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 25',
      code: 'import foo from "./foobar";',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 26',
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 27',
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 28',
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 29',
      code: 'import { foo } from "./ignore.invalid.extension"',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 1 valid 0',
      code: "function x(deepDep) { console.log(deepDep.MY_TERRIBLE_ACTION) } import { deepDep } from './deep-deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 2 valid 0',
      code: "import * as hasDeprecated from './ts-deprecated.ts'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 0',
      code: '\n// @file: ./answer.js\n\n/**\n * this is what you get when you trust a mouse talk show\n * @deprecated need to restart the experiment\n * @returns {Number} nonsense\n */\nexport function multiply(six, nine) {\n  return 42\n}\n',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 1',
      code: '// Deprecated: This is what you get when you trust a mouse talk show, need to\n// restart the experiment.\n//\n// Returns a Number nonsense\nexport function multiply(six, nine) { return 42 }',
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "import { fn } from './deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: "Deprecated: please use 'x' instead.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "import TerribleClass from './deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: this is awful, use NotAsBadClass.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: "import { fn } from './deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/docstyle': ['jsdoc', 'tomdoc'],
      },
      errors: [
        {
          messageId: 'deprecated',
          message: "Deprecated: please use 'x' instead.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "import { fn } from './tomdoc-deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/docstyle': ['tomdoc'],
      },
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: This function is terrible.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: "import TerribleClass from './tomdoc-deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/docstyle': ['tomdoc'],
      },
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: this is awful, use NotAsBadClass.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: "import { MY_TERRIBLE_ACTION } from './tomdoc-deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/docstyle': ['tomdoc'],
      },
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: Please stop sending/handling this action type.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; function shadow(MY_TERRIBLE_ACTION) { console.log(MY_TERRIBLE_ACTION); }",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: "import { MY_TERRIBLE_ACTION, fine } from './deprecated'; console.log(fine)",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(MY_TERRIBLE_ACTION)",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 64,
          endLine: 1,
          endColumn: 82,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(someOther.MY_TERRIBLE_ACTION)",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(MY_TERRIBLE_ACTION.whatever())",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 64,
          endLine: 1,
          endColumn: 82,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(MY_TERRIBLE_ACTION(this, is, the, worst))",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 64,
          endLine: 1,
          endColumn: 82,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: "import Thing from './deprecated-file'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: this module is the worst.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: "import Thing from './deprecated-file'; console.log(other.Thing)",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: this module is the worst.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: "import * as depd from './deprecated'; console.log(depd.MY_TERRIBLE_ACTION)",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 56,
          endLine: 1,
          endColumn: 74,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: "import * as deep from './deep-deprecated'; console.log(deep.deepDep.MY_TERRIBLE_ACTION)",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 69,
          endLine: 1,
          endColumn: 87,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: "import { deepDep } from './deep-deprecated'; console.log(deepDep.MY_TERRIBLE_ACTION)",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 66,
          endLine: 1,
          endColumn: 84,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: "import { deepDep } from './deep-deprecated'; function x(deepNDep) { console.log(deepDep.MY_TERRIBLE_ACTION) }",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 89,
          endLine: 1,
          endColumn: 107,
        },
      ],
    },
    {
      name: 'run 1 invalid 0',
      code: "console.log(MY_TERRIBLE_ACTION); import { MY_TERRIBLE_ACTION } from './deprecated'",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 31,
        },
        {
          messageId: 'deprecated',
          message: 'Deprecated: please stop sending/handling this action type.',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      name: 'run 2 invalid 0',
      code: "import { foo } from './ts-deprecated.ts'; console.log(foo())",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: "Deprecated: don't use this!",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'deprecated',
          message: "Deprecated: don't use this!",
          line: 1,
          column: 55,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    {
      name: 'run 3 invalid 0',
      code: "import { multiply } from './answer'\nfunction whatever(y, z) { return multiply(y, z) }",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      errors: [
        {
          messageId: 'deprecated',
          message: 'Deprecated: need to restart the experiment',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'deprecated',
          message: 'Deprecated: need to restart the experiment',
          line: 2,
          column: 34,
          endLine: 2,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'run 3 invalid 1',
      code: "import { multiply } from './tomdoc-answer'\nfunction whatever(y, z) { return multiply(y, z) }",
      filename: 'fixtures/no-deprecated-rule/consumer.ts',
      settings: {
        'import/docstyle': ['jsdoc', 'tomdoc'],
      },
      errors: [
        {
          messageId: 'deprecated',
          message:
            'Deprecated: This is what you get when you trust a mouse talk show, need to restart the experiment.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'deprecated',
          message:
            'Deprecated: This is what you get when you trust a mouse talk show, need to restart the experiment.',
          line: 2,
          column: 34,
          endLine: 2,
          endColumn: 42,
        },
      ],
    },
  ],
});
