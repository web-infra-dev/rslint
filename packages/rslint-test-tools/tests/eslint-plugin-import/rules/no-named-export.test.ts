// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-named-export', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'module.export.foo = function () {}',
    },
    {
      name: 'run 0 valid 1',
      code: 'module.export.foo = function () {}',
    },
    {
      name: 'run 0 valid 2',
      code: 'export default function bar() {};',
    },
    {
      name: 'run 0 valid 3',
      code: 'let foo; export { foo as default }',
    },
    {
      name: 'run 0 valid 4',
      code: "import * as foo from './foo';",
    },
    {
      name: 'run 0 valid 5',
      code: "import foo from './foo';",
    },
    {
      name: 'run 0 valid 6',
      code: "import {default as foo} from './foo';",
    },
    {
      name: 'run 0 valid 7',
      code: 'let foo; export { foo as "default" }',
    },
    {
      name: 'run 0 valid 8',
      code: "// good1.js\n\n// There is only a single module export and it's a default export.\nexport default 'bar';",
    },
    {
      name: 'run 0 valid 9',
      code: "// good2.js\n\n// There is only a single module export and it's a default export.\nconst foo = 'foo';\nexport { foo as default }",
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "\n        export const foo = 'foo';\n        export const bar = 'bar';\n      ",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 34,
        },
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "\n        export const foo = 'foo';\n        export default bar;",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "\n        export const foo = 'foo';\n        export function bar() {};\n      ",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 34,
        },
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: "export const foo = 'foo';",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "\n        const foo = 'foo';\n        export { foo };\n      ",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'let foo, bar; export { foo, bar }',
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'export const { foo, bar } = item;',
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'export const { foo, bar: baz } = item;',
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'export const { foo: { bar, baz } } = item;',
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: '\n        let item;\n        export const foo = item;\n        export { item };\n      ',
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 33,
        },
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: "export * from './foo';",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'export const { foo } = { foo: "bar" };',
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: 'export const { foo: { bar } } = { foo: { bar: "baz" } };',
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'export { a, b } from "foo.js"',
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'export type UserId = number;',
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: "// bad1.js\n\n// There is only a single module export and it's a named export.\nexport const foo = 'foo';",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: "// bad2.js\n\n// There is more than one named export in the module.\nexport const foo = 'foo';\nexport const bar = 'bar';",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 26,
        },
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: "// bad3.js\n\n// There is more than one named export in the module.\nconst foo = 'foo';\nconst bar = 'bar';\nexport { foo, bar }",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: "// bad4.js\n\n// There is more than one named export in the module.\nexport * from './other-module'",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 19',
      code: "// bad5.js\n\n// There is a default and a named export.\nexport const foo = 'foo';\nconst bar = 'bar';\nexport default 'bar';",
      errors: [
        {
          messageId: '',
          message: 'Named exports are not allowed.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 26,
        },
      ],
    },
  ],
});
