import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-restricted-exports', {
  valid: [
    // nothing configured
    'export var a;',
    'export function a() {}',
    'export class A {}',
    'var a; export { a };',
    'export { a } from "foo";',
    { code: 'export var a;', options: {} },

    // not a restricted name
    { code: 'export var a;', options: { restrictedNamedExports: ['x'] } },
    {
      code: 'export function a() {}',
      options: { restrictedNamedExports: ['x'] },
    },
    { code: 'export class A {}', options: { restrictedNamedExports: ['x'] } },
    {
      code: 'var a; export { a };',
      options: { restrictedNamedExports: ['x'] },
    },
    {
      code: "export { a } from 'foo';",
      options: { restrictedNamedExports: ['x'] },
    },
    {
      code: "export { '' } from 'foo';",
      options: { restrictedNamedExports: ['undefined'] },
    },

    // does not mistakenly disallow non-exported names that appear in named export declarations
    {
      code: 'export var b = a;',
      options: { restrictedNamedExports: ['a'] },
    },
    {
      code: 'export var { a: b } = {};',
      options: { restrictedNamedExports: ['a'] },
    },
    {
      code: 'var a; export { a as b };',
      options: { restrictedNamedExports: ['a'] },
    },

    // does not check source in re-export declarations
    {
      code: "export { b } from 'a';",
      options: { restrictedNamedExports: ['a'] },
    },
    {
      code: "export * as b from 'a';",
      options: { restrictedNamedExports: ['a'] },
    },

    // does not check non-export declarations
    { code: 'var a;', options: { restrictedNamedExports: ['a'] } },
    {
      code: "import { a } from 'foo';",
      options: { restrictedNamedExports: ['a'] },
    },
    {
      code: 'var setSomething; export { setSomething };',
      options: { restrictedNamedExportsPattern: '^get' },
    },

    // does not check re-export all declarations
    {
      code: "export * from 'foo';",
      options: { restrictedNamedExports: ['a'] },
    },

    // does not mistakenly disallow identifiers in export default declarations
    // (a default export will export "default" name)
    {
      code: 'export default a;',
      options: { restrictedNamedExports: ['a'] },
    },
    {
      code: 'export default function a() {}',
      options: { restrictedNamedExports: ['a'] },
    },

    // by design, restricted name "default" does not apply to default export
    // declarations, although they do export the "default" name.
    {
      code: 'export default 1;',
      options: { restrictedNamedExports: ['default'] },
    },

    // "default" does not disallow re-exporting a renamed default export from another module
    {
      code: "export { default as a } from 'foo';",
      options: { restrictedNamedExports: ['default'] },
    },

    // restrictDefaultExports.direct option
    {
      code: 'export default foo;',
      options: { restrictDefaultExports: { direct: false } },
    },

    // restrictDefaultExports.named option
    {
      code: 'const foo = 123;\nexport { foo as default };',
      options: { restrictDefaultExports: { named: false } },
    },

    // restrictDefaultExports.defaultFrom option
    {
      code: "export { default } from 'mod';",
      options: { restrictDefaultExports: { defaultFrom: false } },
    },

    // restrictDefaultExports.namedFrom option
    {
      code: "export { foo as default } from 'mod';",
      options: { restrictDefaultExports: { namedFrom: false } },
    },

    // restrictDefaultExports.namespaceFrom option
    {
      code: "export * as default from 'mod';",
      options: { restrictDefaultExports: { namespaceFrom: false } },
    },
  ],

  invalid: [
    {
      code: 'export function someFunction() {}',
      options: { restrictedNamedExports: ['someFunction'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 17 }],
    },

    // basic tests
    {
      code: 'export var a;',
      options: { restrictedNamedExports: ['a'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 12 }],
    },
    {
      code: 'export function a() {}',
      options: { restrictedNamedExports: ['a'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 17 }],
    },
    {
      code: 'export class A {}',
      options: { restrictedNamedExports: ['A'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 14 }],
    },
    {
      code: 'let a; export { a };',
      options: { restrictedNamedExports: ['a'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 17 }],
    },
    {
      code: "export { a } from 'foo';",
      options: { restrictedNamedExports: ['a'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 10 }],
    },

    // string literals
    {
      code: "let a; export { a as 'a' };",
      options: { restrictedNamedExports: ['a'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 22 }],
    },
    {
      code: "export { '' } from 'foo';",
      options: { restrictedNamedExports: [''] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 10 }],
    },

    // destructuring
    {
      code: 'export var [a] = [];',
      options: { restrictedNamedExports: ['a'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 13 }],
    },
    {
      code: 'export let { b: { c: a = d } = e } = {};',
      options: { restrictedNamedExports: ['a'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 22 }],
    },

    // reports the correct identifier node in the case of a redeclaration
    {
      code: 'var a; export var a;',
      options: { restrictedNamedExports: ['a'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 19 }],
    },

    // multiple invalid in the same declaration
    {
      code: 'export const b = 1, a = 2;',
      options: { restrictedNamedExports: ['a', 'b'] },
      errors: [
        { messageId: 'restrictedNamed', line: 1, column: 14 },
        { messageId: 'restrictedNamed', line: 1, column: 21 },
      ],
    },

    // restrictedNamedExportsPattern
    {
      code: 'var getSomething; export { getSomething };',
      options: { restrictedNamedExportsPattern: 'get*' },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 28 }],
    },

    // reports "default" in named export declarations (when configured)
    {
      code: 'var a; export { a as default };',
      options: { restrictedNamedExports: ['default'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 22 }],
    },
    {
      code: "export { default } from 'foo';",
      options: { restrictedNamedExports: ['default'] },
      errors: [{ messageId: 'restrictedNamed', line: 1, column: 10 }],
    },

    // restrictDefaultExports.direct option
    {
      code: 'export default foo;',
      options: { restrictDefaultExports: { direct: true } },
      errors: [{ messageId: 'restrictedDefault', line: 1, column: 1 }],
    },
    {
      code: 'export default function foo() {}',
      options: { restrictDefaultExports: { direct: true } },
      errors: [{ messageId: 'restrictedDefault', line: 1, column: 1 }],
    },

    // restrictDefaultExports.named option
    {
      code: 'const foo = 123;\nexport { foo as default };',
      options: { restrictDefaultExports: { named: true } },
      errors: [{ messageId: 'restrictedDefault', line: 2, column: 17 }],
    },

    // restrictDefaultExports.defaultFrom option
    {
      code: "export { default } from 'mod';",
      options: { restrictDefaultExports: { defaultFrom: true } },
      errors: [{ messageId: 'restrictedDefault', line: 1, column: 10 }],
    },

    // restrictDefaultExports.namedFrom option
    {
      code: "export { foo as default } from 'mod';",
      options: { restrictDefaultExports: { namedFrom: true } },
      errors: [{ messageId: 'restrictedDefault', line: 1, column: 17 }],
    },

    // restrictDefaultExports.namespaceFrom option
    {
      code: "export * as default from 'mod';",
      options: { restrictDefaultExports: { namespaceFrom: true } },
      errors: [{ messageId: 'restrictedDefault', line: 1, column: 13 }],
    },
  ],
});
