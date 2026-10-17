import { RuleTester } from '../rule-tester.js';
import { testFixturePath } from '../utils.js';

// All eslint-plugin-import v2.32.0 no-deprecated groups, SYNTAX_CASES and docs.
// Go tests assert complete ranges and the native message ID.
const filename = testFixturePath('no-deprecated-rule/consumer.ts');
const ruleTester = new RuleTester();

// no-deprecated
ruleTester.run('no-deprecated', null as never, {
  valid: [
    {
      code: "import { x } from './fake' ",
    },
    {
      code: "import bar from './bar'",
    },
    {
      code: "import { fine } from './deprecated'",
    },
    {
      code: "import { _undocumented } from './deprecated'",
    },
    {
      code: "import { fn } from './deprecated'",
      settings: {
        'import/docstyle': ['tomdoc'],
      },
    },
    {
      code: "import { fine } from './tomdoc-deprecated'",
      settings: {
        'import/docstyle': ['tomdoc'],
      },
    },
    {
      code: "import { _undocumented } from './tomdoc-deprecated'",
      settings: {
        'import/docstyle': ['tomdoc'],
      },
    },
    {
      code: "import * as depd from './deprecated'",
    },
    {
      code: "import * as depd from './deprecated'; console.log(depd.fine())",
    },
    {
      code: "import { deepDep } from './deep-deprecated'",
    },
    {
      code: "import { deepDep } from './deep-deprecated'; console.log(deepDep.fine())",
    },
    {
      code: "import { deepDep } from './deep-deprecated'; function x(deepDep) { console.log(deepDep.MY_TERRIBLE_ACTION) }",
    },
    {
      code: 'for (let { foo, bar } of baz) {}',
    },
    {
      code: 'for (let [ foo, bar ] of baz) {}',
    },
    {
      code: 'const { x, y } = bar',
    },
    {
      code: 'const { x, y, ...z } = bar',
    },
    {
      code: 'let x; export { x }',
    },
    {
      code: 'let x; export { x as y }',
    },
    {
      code: 'export const x = null',
    },
    {
      code: 'export var x = null',
    },
    {
      code: 'export let x = null',
    },
    {
      code: 'export default x',
    },
    {
      code: 'export default class x {}',
    },
    {
      code: 'import json from "./data.json"',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      code: 'import foo from "./foobar.json";',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      code: 'import foo from "./foobar";',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
    },
    {
      code: 'import { foo } from "./ignore.invalid.extension"',
    },
  ].map((c) => ({ ...c, filename })),
  invalid: [
    {
      code: "import { fn } from './deprecated'",
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
      code: "import TerribleClass from './deprecated'",
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
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'",
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
      code: "import { fn } from './deprecated'",
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
      settings: {
        'import/docstyle': ['jsdoc', 'tomdoc'],
      },
    },
    {
      code: "import { fn } from './tomdoc-deprecated'",
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
      settings: {
        'import/docstyle': ['tomdoc'],
      },
    },
    {
      code: "import TerribleClass from './tomdoc-deprecated'",
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
      settings: {
        'import/docstyle': ['tomdoc'],
      },
    },
    {
      code: "import { MY_TERRIBLE_ACTION } from './tomdoc-deprecated'",
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
      settings: {
        'import/docstyle': ['tomdoc'],
      },
    },
    {
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; function shadow(MY_TERRIBLE_ACTION) { console.log(MY_TERRIBLE_ACTION); }",
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
      code: "import { MY_TERRIBLE_ACTION, fine } from './deprecated'; console.log(fine)",
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
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(MY_TERRIBLE_ACTION)",
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
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(someOther.MY_TERRIBLE_ACTION)",
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
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(MY_TERRIBLE_ACTION.whatever())",
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
      code: "import { MY_TERRIBLE_ACTION } from './deprecated'; console.log(MY_TERRIBLE_ACTION(this, is, the, worst))",
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
      code: "import Thing from './deprecated-file'",
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
      code: "import Thing from './deprecated-file'; console.log(other.Thing)",
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
      code: "import * as depd from './deprecated'; console.log(depd.MY_TERRIBLE_ACTION)",
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
      code: "import * as deep from './deep-deprecated'; console.log(deep.deepDep.MY_TERRIBLE_ACTION)",
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
      code: "import { deepDep } from './deep-deprecated'; console.log(deepDep.MY_TERRIBLE_ACTION)",
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
      code: "import { deepDep } from './deep-deprecated'; function x(deepNDep) { console.log(deepDep.MY_TERRIBLE_ACTION) }",
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
  ].map((c) => ({ ...c, filename })),
});
// Retained upstream gap: import './malformed.js' expects one parser-error report.
// The shared module map does not translate dependency parser diagnostics.

// no-deprecated: hoisting
ruleTester.run('no-deprecated', null as never, {
  valid: [
    {
      code: "function x(deepDep) { console.log(deepDep.MY_TERRIBLE_ACTION) } import { deepDep } from './deep-deprecated'",
    },
  ].map((c) => ({ ...c, filename })),
  invalid: [
    {
      code: "console.log(MY_TERRIBLE_ACTION); import { MY_TERRIBLE_ACTION } from './deprecated'",
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
  ].map((c) => ({ ...c, filename })),
});

// typescript
ruleTester.run('no-deprecated', null as never, {
  valid: [
    {
      code: "import * as hasDeprecated from './ts-deprecated.ts'",
    },
  ].map((c) => ({ ...c, filename })),
  invalid: [
    {
      code: "import { foo } from './ts-deprecated.ts'; console.log(foo())",
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
  ].map((c) => ({ ...c, filename })),
});

// documentation
ruleTester.run('no-deprecated', null as never, {
  valid: [
    {
      code: '\n// @file: ./answer.js\n\n/**\n * this is what you get when you trust a mouse talk show\n * @deprecated need to restart the experiment\n * @returns {Number} nonsense\n */\nexport function multiply(six, nine) {\n  return 42\n}\n',
    },
    {
      code: '// Deprecated: This is what you get when you trust a mouse talk show, need to\n// restart the experiment.\n//\n// Returns a Number nonsense\nexport function multiply(six, nine) { return 42 }',
    },
  ].map((c) => ({ ...c, filename })),
  invalid: [
    {
      code: "import { multiply } from './answer'\nfunction whatever(y, z) { return multiply(y, z) }",
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
      code: "import { multiply } from './tomdoc-answer'\nfunction whatever(y, z) { return multiply(y, z) }",
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
      settings: {
        'import/docstyle': ['jsdoc', 'tomdoc'],
      },
    },
  ].map((c) => ({ ...c, filename })),
});
