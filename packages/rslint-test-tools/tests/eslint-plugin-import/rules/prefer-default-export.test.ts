// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('prefer-default-export', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: "\n        export const foo = 'foo';\n        export const bar = 'bar';",
    },
    {
      name: 'run 0 valid 1',
      code: '\n        export default function bar() {};',
    },
    {
      name: 'run 0 valid 2',
      code: "\n        export const foo = 'foo';\n        export function bar() {};",
    },
    {
      name: 'run 0 valid 3',
      code: "\n        export const foo = 'foo';\n        export default bar;",
    },
    {
      name: 'run 0 valid 4',
      code: '\n        let foo, bar;\n        export { foo, bar }',
    },
    {
      name: 'run 0 valid 5',
      code: '\n        export const { foo, bar } = item;',
    },
    {
      name: 'run 0 valid 6',
      code: '\n        export const { foo, bar: baz } = item;',
    },
    {
      name: 'run 0 valid 7',
      code: '\n        export const { foo: { bar, baz } } = item;',
    },
    {
      name: 'run 0 valid 8',
      code: '\n        export const [a, b] = item;',
    },
    {
      name: 'run 0 valid 9',
      code: '\n        let item;\n        export const foo = item;\n        export { item };',
    },
    {
      name: 'run 0 valid 10',
      code: '\n        let foo;\n        export { foo as default }',
    },
    {
      name: 'run 0 valid 11',
      code: "\n        export * from './foo';",
    },
    {
      name: 'run 0 valid 12',
      code: "\n        import * as foo from './foo';",
    },
    {
      name: 'run 0 valid 13',
      code: 'export type UserId = number;',
    },
    {
      name: 'run 0 valid 14',
      code: 'export { a, b } from "foo.js"',
    },
    {
      name: 'run 0 valid 15',
      code: '\n        export const [CounterProvider,, withCounter] = func();;\n      ',
    },
    {
      name: 'run 0 valid 16',
      code: 'let foo; export { foo as "default" };',
    },
    {
      name: 'run 0 valid 17',
      code: '\n          export default function bar() {};',
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: "\n              export const foo = 'foo';\n              export const bar = 'bar';\n              export default 42;",
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: '\n            export default a = 2;',
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 20',
      code: '\n            export const a = 2;\n            export default function foo() {};',
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 21',
      code: '\n          export const a = 5;\n          export function bar(){};\n          let foo;\n          export { foo as default }',
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 22',
      code: "\n          export * from './foo';",
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 23',
      code: "\n            import * as foo from './foo';",
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 24',
      code: 'const a = 5;',
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 25',
      code: 'export const a = 4; let foo; export { foo as "default" };',
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 26',
      code: '\n            export type foo = string;\n            export type bar = number;\n            /* @typescript-eslint/parser */\n          ',
    },
    {
      name: 'run 0 valid 27',
      code: '\n            export type foo = string;\n            export type bar = number;\n            /* @typescript-eslint/parser */\n          ',
    },
    {
      name: 'run 0 valid 28',
      code: 'export type foo = string /* @typescript-eslint/parser*/',
    },
    {
      name: 'run 0 valid 29',
      code: 'export interface foo { bar: string; } /* @typescript-eslint/parser*/',
    },
    {
      name: 'run 0 valid 30',
      code: 'export interface foo { bar: string; }; export function goo() {} /* @typescript-eslint/parser*/',
    },
    {
      name: 'run 0 valid 31',
      code: "// good1.js\n\n// There is a default export.\nexport const foo = 'foo';\nconst bar = 'bar';\nexport default bar;",
    },
    {
      name: 'run 0 valid 32',
      code: "// good2.js\n\n// There is more than one named export in the module.\nexport const foo = 'foo';\nexport const bar = 'bar';",
    },
    {
      name: 'run 0 valid 33',
      code: "// good3.js\n\n// There is more than one named export in the module\nconst foo = 'foo';\nconst bar = 'bar';\nexport { foo, bar }",
    },
    {
      name: 'run 0 valid 34',
      code: "// good4.js\n\n// There is a default export.\nconst foo = 'foo';\nexport { foo as default }",
    },
    {
      name: 'run 0 valid 35',
      code: "// export-star.js\n\n// Any batch export will disable this rule. The remote module is not inspected.\nexport * from './other-module'",
    },
    {
      name: 'run 0 valid 36',
      code: '// good1.js\n\n//has default export\nexport default function bar() {};',
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 37',
      code: '// good2.js\n\n// has default export\nlet foo;\nexport { foo as default }',
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 38',
      code: '// good3.js\n\n//contains multiple exports AND default export\nexport const a = 5;\nexport function bar(){};\nlet foo;\nexport { foo as default }',
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 39',
      code: "// good4.js\n\n// does not contain any exports => file is not checked by the rule\nimport * as foo from './foo';﻿",
      options: [
        {
          target: 'any',
        },
      ],
    },
    {
      name: 'run 0 valid 40',
      code: "// export-star.js\n\n// Any batch export will disable this rule. The remote module is not inspected.\nexport * from './other-module'",
      options: [
        {
          target: 'any',
        },
      ],
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: '\n        export function bar() {};',
      errors: [
        {
          messageId: '',
          message: 'Prefer default export on a file with single export.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "\n        export const foo = 'foo';",
      errors: [
        {
          messageId: '',
          message: 'Prefer default export on a file with single export.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "\n        const foo = 'foo';\n        export { foo };",
      errors: [
        {
          messageId: '',
          message: 'Prefer default export on a file with single export.',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: '\n        export const { foo } = { foo: "bar" };',
      errors: [
        {
          messageId: '',
          message: 'Prefer default export on a file with single export.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 47,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: '\n        export const { foo: { bar } } = { foo: { bar: "baz" } };',
      errors: [
        {
          messageId: '',
          message: 'Prefer default export on a file with single export.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 65,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: '\n        export const [a] = ["foo"]',
      errors: [
        {
          messageId: '',
          message: 'Prefer default export on a file with single export.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: "\n        export const foo = 'foo';\n        export const bar = 'bar';",
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: "\n        export const foo = 'foo';\n        export function bar() {};",
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: '\n        let foo, bar;\n        export { foo, bar }',
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: '\n        let item;\n        export const foo = item;\n        export { item };',
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'export { a, b } from "foo.js"',
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: "\n        const foo = 'foo';\n        export { foo };",
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: '\n        export const { foo } = { foo: "bar" };',
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 47,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: '\n        export const { foo: { bar } } = { foo: { bar: "baz" } };',
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 65,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: "// bad.js\n\n// There is only a single module export and it's a named export.\nexport const foo = 'foo';\n",
      errors: [
        {
          messageId: '',
          message: 'Prefer default export on a file with single export.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: "// bad1.js\n\n//has 2 named exports, but no default export\nexport const foo = 'foo';\nexport const bar = 'bar';",
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: '// bad2.js\n\n// does not have default export\nlet foo, bar;\nexport { foo, bar }',
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: '// bad3.js\n\n// does not have default export\nexport { a, b } from "foo.js"﻿',
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: '// bad4.js\n\n// does not have default export\nlet item;\nexport const foo = item;\nexport { item };',
      options: [
        {
          target: 'any',
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            'Prefer default export to be present on every file that has export.',
          line: 6,
          column: 10,
          endLine: 6,
          endColumn: 14,
        },
      ],
    },
  ],
});
