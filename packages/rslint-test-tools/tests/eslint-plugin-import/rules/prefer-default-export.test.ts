// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/prefer-default-export.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/prefer-default-export.md
// Parser variants share cases. Go tests assert full ranges, message IDs and absence of edits.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
ruleTester.run('prefer-default-export', undefined, {
  valid: [
    // Single
    {
      code: "\n        export const foo = 'foo';\n        export const bar = 'bar';",
    },
    { code: '\n        export default function bar() {};' },
    {
      code: "\n        export const foo = 'foo';\n        export function bar() {};",
    },
    {
      code: "\n        export const foo = 'foo';\n        export default bar;",
    },
    { code: '\n        let foo, bar;\n        export { foo, bar }' },
    { code: '\n        export const { foo, bar } = item;' },
    { code: '\n        export const { foo, bar: baz } = item;' },
    { code: '\n        export const { foo: { bar, baz } } = item;' },
    { code: '\n        export const [a, b] = item;' },
    {
      code: '\n        let item;\n        export const foo = item;\n        export { item };',
    },
    { code: '\n        let foo;\n        export { foo as default }' },
    { code: "\n        export * from './foo';" },
    // SKIP: unsupported Babel default re-export proposal: "export Memory, { MemoryValue } from './Memory'"
    { code: "\n        import * as foo from './foo';" },
    { code: 'export type UserId = number;' },
    // SKIP: unsupported Babel default re-export proposal: "export default from \"foo.js\""
    { code: 'export { a, b } from "foo.js"' },
    {
      code: '\n        export const [CounterProvider,, withCounter] = func();;\n      ',
    },
    { code: 'let foo; export { foo as "default" };' },
    // Any
    {
      code: '\n          export default function bar() {};',
      options: [{ target: 'any' }],
    },
    {
      code: "\n              export const foo = 'foo';\n              export const bar = 'bar';\n              export default 42;",
      options: [{ target: 'any' }],
    },
    {
      code: '\n            export default a = 2;',
      options: [{ target: 'any' }],
    },
    {
      code: '\n            export const a = 2;\n            export default function foo() {};',
      options: [{ target: 'any' }],
    },
    {
      code: '\n          export const a = 5;\n          export function bar(){};\n          let foo;\n          export { foo as default }',
      options: [{ target: 'any' }],
    },
    {
      code: "\n          export * from './foo';",
      options: [{ target: 'any' }],
    },
    // SKIP: unsupported Babel default re-export proposal: "export Memory, { MemoryValue } from './Memory'"
    {
      code: "\n            import * as foo from './foo';",
      options: [{ target: 'any' }],
    },
    { code: 'const a = 5;', options: [{ target: 'any' }] },
    {
      code: 'export const a = 4; let foo; export { foo as "default" };',
      options: [{ target: 'any' }],
    },
    // TypeScript
    {
      code: '\n            export type foo = string;\n            export type bar = number;\n            /* @typescript-eslint/parser */\n          ',
    },
    {
      code: '\n            export type foo = string;\n            export type bar = number;\n            /* @typescript-eslint/parser */\n          ',
    },
    { code: 'export type foo = string /* @typescript-eslint/parser*/' },
    {
      code: 'export interface foo { bar: string; } /* @typescript-eslint/parser*/',
    },
    {
      code: 'export interface foo { bar: string; }; export function goo() {} /* @typescript-eslint/parser*/',
    },
    // DocsSingle
    {
      code: "// good1.js\n\n// There is a default export.\nexport const foo = 'foo';\nconst bar = 'bar';\nexport default bar;",
    },
    {
      code: "// good2.js\n\n// There is more than one named export in the module.\nexport const foo = 'foo';\nexport const bar = 'bar';",
    },
    {
      code: "// good3.js\n\n// There is more than one named export in the module\nconst foo = 'foo';\nconst bar = 'bar';\nexport { foo, bar }",
    },
    {
      code: "// good4.js\n\n// There is a default export.\nconst foo = 'foo';\nexport { foo as default }",
    },
    {
      code: "// export-star.js\n\n// Any batch export will disable this rule. The remote module is not inspected.\nexport * from './other-module'",
    },
    // DocsAny
    {
      code: '// good1.js\n\n//has default export\nexport default function bar() {};',
      options: [{ target: 'any' }],
    },
    {
      code: '// good2.js\n\n// has default export\nlet foo;\nexport { foo as default }',
      options: [{ target: 'any' }],
    },
    {
      code: '// good3.js\n\n//contains multiple exports AND default export\nexport const a = 5;\nexport function bar(){};\nlet foo;\nexport { foo as default }',
      options: [{ target: 'any' }],
    },
    {
      code: "// good4.js\n\n// does not contain any exports => file is not checked by the rule\nimport * as foo from './foo';\ufeff",
      options: [{ target: 'any' }],
    },
    {
      code: "// export-star.js\n\n// Any batch export will disable this rule. The remote module is not inspected.\nexport * from './other-module'",
      options: [{ target: 'any' }],
    },
  ],
  invalid: [
    // Single
    {
      code: '\n        export function bar() {};',
      errors: ['Prefer default export on a file with single export.'],
    },
    {
      code: "\n        export const foo = 'foo';",
      errors: ['Prefer default export on a file with single export.'],
    },
    {
      code: "\n        const foo = 'foo';\n        export { foo };",
      errors: ['Prefer default export on a file with single export.'],
    },
    {
      code: '\n        export const { foo } = { foo: "bar" };',
      errors: ['Prefer default export on a file with single export.'],
    },
    {
      code: '\n        export const { foo: { bar } } = { foo: { bar: "baz" } };',
      errors: ['Prefer default export on a file with single export.'],
    },
    {
      code: '\n        export const [a] = ["foo"]',
      errors: ['Prefer default export on a file with single export.'],
    },
    // Any
    {
      code: "\n        export const foo = 'foo';\n        export const bar = 'bar';",
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: "\n        export const foo = 'foo';\n        export function bar() {};",
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: '\n        let foo, bar;\n        export { foo, bar }',
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: '\n        let item;\n        export const foo = item;\n        export { item };',
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: 'export { a, b } from "foo.js"',
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: "\n        const foo = 'foo';\n        export { foo };",
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: '\n        export const { foo } = { foo: "bar" };',
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: '\n        export const { foo: { bar } } = { foo: { bar: "baz" } };',
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    // DocsSingle
    {
      code: "// bad.js\n\n// There is only a single module export and it's a named export.\nexport const foo = 'foo';\n",
      errors: ['Prefer default export on a file with single export.'],
    },
    // DocsAny
    {
      code: "// bad1.js\n\n//has 2 named exports, but no default export\nexport const foo = 'foo';\nexport const bar = 'bar';",
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: '// bad2.js\n\n// does not have default export\nlet foo, bar;\nexport { foo, bar }',
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: '// bad3.js\n\n// does not have default export\nexport { a, b } from "foo.js"\ufeff',
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
    {
      code: '// bad4.js\n\n// does not have default export\nlet item;\nexport const foo = item;\nexport { item };',
      options: [{ target: 'any' }],
      errors: [
        'Prefer default export to be present on every file that has export.',
      ],
    },
  ],
});
