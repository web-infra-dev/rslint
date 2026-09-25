import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-anonymous-default-export.js
// Complete upstream mirror. Exact ranges and absence of edits are asserted in Go.
const ruleTester = new RuleTester();

ruleTester.run('no-anonymous-default-export', null as never, {
  valid: [
    // Exports with identifiers. The docs typo `class MyClass() {}` is covered below as `class MyClass {}`.
    {
      code: 'const foo = 123\nexport default foo',
    },
    {
      code: 'export default function foo() {}',
    },
    {
      code: 'export default class MyClass {}',
    },
    // Allow each forbidden type with its option.
    {
      code: 'export default []',
      options: [
        {
          allowArray: true,
        },
      ],
    },
    {
      code: 'export default () => {}',
      options: [
        {
          allowArrowFunction: true,
        },
      ],
    },
    {
      code: 'export default class {}',
      options: [
        {
          allowAnonymousClass: true,
        },
      ],
    },
    {
      code: 'export default function() {}',
      options: [
        {
          allowAnonymousFunction: true,
        },
      ],
    },
    {
      code: 'export default 123',
      options: [
        {
          allowLiteral: true,
        },
      ],
    },
    {
      code: "export default 'foo'",
      options: [
        {
          allowLiteral: true,
        },
      ],
    },
    {
      code: 'export default `foo`',
      options: [
        {
          allowLiteral: true,
        },
      ],
    },
    {
      code: 'export default {}',
      options: [
        {
          allowObject: true,
        },
      ],
    },
    {
      code: 'export default foo(bar)',
      options: [
        {
          allowCallExpression: true,
        },
      ],
    },
    {
      code: 'export default new Foo()',
      options: [
        {
          allowNew: true,
        },
      ],
    },
    // Multiple options.
    {
      code: 'export default 123',
      options: [
        {
          allowLiteral: true,
          allowObject: true,
        },
      ],
    },
    {
      code: 'export default {}',
      options: [
        {
          allowLiteral: true,
          allowObject: true,
        },
      ],
    },
    // Unrelated export syntaxes, including the ESLint >= 8.7 arbitrary export name case.
    {
      code: "export * from 'foo'",
    },
    {
      code: 'const foo = 123\nexport { foo }',
    },
    {
      code: 'const foo = 123\nexport { foo as default }',
    },
    {
      code: 'const foo = 123\nexport { foo as "default" }',
    },
    // Call expressions are allowed by default for backwards compatibility.
    {
      code: 'export default foo(bar)',
    },
    // All 18 SYNTAX_CASES from v2.32.0/tests/src/utils.js; object rest uses the native parser.
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
    // Documentation example with a space before the function parameters. Other docs examples match cases above.
    {
      code: 'export default function () {}',
      options: [
        {
          allowAnonymousFunction: true,
        },
      ],
    },
  ],
  invalid: [
    // Every upstream invalid case. Upstream uses message text without message IDs.
    {
      code: 'export default []',
      errors: [
        {
          message:
            'Assign array to a variable before exporting as module default',
        },
      ],
    },
    {
      code: 'export default () => {}',
      errors: [
        {
          message:
            'Assign arrow function to a variable before exporting as module default',
        },
      ],
    },
    {
      code: 'export default class {}',
      errors: [
        {
          message: 'Unexpected default export of anonymous class',
        },
      ],
    },
    {
      code: 'export default function() {}',
      errors: [
        {
          message: 'Unexpected default export of anonymous function',
        },
      ],
    },
    {
      code: 'export default 123',
      errors: [
        {
          message:
            'Assign literal to a variable before exporting as module default',
        },
      ],
    },
    {
      code: "export default 'foo'",
      errors: [
        {
          message:
            'Assign literal to a variable before exporting as module default',
        },
      ],
    },
    {
      code: 'export default `foo`',
      errors: [
        {
          message:
            'Assign literal to a variable before exporting as module default',
        },
      ],
    },
    {
      code: 'export default {}',
      errors: [
        {
          message:
            'Assign object to a variable before exporting as module default',
        },
      ],
    },
    {
      code: 'export default foo(bar)',
      options: [
        {
          allowCallExpression: false,
        },
      ],
      errors: [
        {
          message:
            'Assign call result to a variable before exporting as module default',
        },
      ],
    },
    {
      code: 'export default new Foo()',
      errors: [
        {
          message:
            'Assign instance to a variable before exporting as module default',
        },
      ],
    },
    // An unrelated option does not suppress the diagnostic.
    {
      code: 'export default 123',
      options: [
        {
          allowObject: true,
        },
      ],
      errors: [
        {
          message:
            'Assign literal to a variable before exporting as module default',
        },
      ],
    },
    // Documentation example; the other failing examples are already covered above.
    {
      code: 'export default function () {}',
      errors: [
        {
          message: 'Unexpected default export of anonymous function',
        },
      ],
    },
  ],
});
