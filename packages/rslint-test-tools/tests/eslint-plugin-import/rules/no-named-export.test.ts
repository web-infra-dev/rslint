import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-named-export.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-named-export.md
const message = 'Named exports are not allowed.';

const ruleTester = new RuleTester();
ruleTester.run('no-named-export', null as never, {
  valid: [
    // Upstream tests.
    // Go also runs this case with sourceType: script; this wrapper uses module mode.
    {
      code: 'module.export.foo = function () {}',
    },
    {
      code: 'module.export.foo = function () {}',
    },
    {
      code: 'export default function bar() {};',
    },
    {
      code: 'let foo; export { foo as default }',
    },
    // SKIP: Babel default re-export proposals are not supported by the parser.
    // export default from "foo.js"
    {
      code: "import * as foo from './foo';",
    },
    {
      code: "import foo from './foo';",
    },
    {
      code: "import {default as foo} from './foo';",
    },
    {
      code: 'let foo; export { foo as "default" }',
    },
    // Upstream documentation.
    {
      code: "// good1.js\n\n// There is only a single module export and it's a default export.\nexport default 'bar';",
    },
    {
      code: "// good2.js\n\n// There is only a single module export and it's a default export.\nconst foo = 'foo';\nexport { foo as default }",
    },
    // SKIP: Babel default re-export proposals are not supported by the parser.
    // good3.js: export default from './other-module';
  ],
  invalid: [
    // Upstream tests.
    {
      code: "\n        export const foo = 'foo';\n        export const bar = 'bar';\n      ",
      errors: [message, message],
    },
    {
      code: "\n        export const foo = 'foo';\n        export default bar;",
      errors: [message],
    },
    {
      code: "\n        export const foo = 'foo';\n        export function bar() {};\n      ",
      errors: [message, message],
    },
    {
      code: "export const foo = 'foo';",
      errors: [message],
    },
    {
      code: "\n        const foo = 'foo';\n        export { foo };\n      ",
      errors: [message],
    },
    {
      code: 'let foo, bar; export { foo, bar }',
      errors: [message],
    },
    {
      code: 'export const { foo, bar } = item;',
      errors: [message],
    },
    {
      code: 'export const { foo, bar: baz } = item;',
      errors: [message],
    },
    {
      code: 'export const { foo: { bar, baz } } = item;',
      errors: [message],
    },
    {
      code: '\n        let item;\n        export const foo = item;\n        export { item };\n      ',
      errors: [message, message],
    },
    {
      code: "export * from './foo';",
      errors: [message],
    },
    {
      code: 'export const { foo } = { foo: "bar" };',
      errors: [message],
    },
    {
      code: 'export const { foo: { bar } } = { foo: { bar: "baz" } };',
      errors: [message],
    },
    {
      code: 'export { a, b } from "foo.js"',
      errors: [message],
    },
    {
      code: 'export type UserId = number;',
      errors: [message],
    },
    // SKIP: Babel default re-export proposals are not supported by the parser.
    // export foo from "foo.js"
    // SKIP: Babel default re-export proposals are not supported by the parser.
    // export Memory, { MemoryValue } from './Memory'
    // Upstream documentation.
    {
      code: "// bad1.js\n\n// There is only a single module export and it's a named export.\nexport const foo = 'foo';",
      errors: [message],
    },
    {
      code: "// bad2.js\n\n// There is more than one named export in the module.\nexport const foo = 'foo';\nexport const bar = 'bar';",
      errors: [message, message],
    },
    {
      code: "// bad3.js\n\n// There is more than one named export in the module.\nconst foo = 'foo';\nconst bar = 'bar';\nexport { foo, bar }",
      errors: [message],
    },
    {
      code: "// bad4.js\n\n// There is more than one named export in the module.\nexport * from './other-module'",
      errors: [message],
    },
    {
      code: "// bad5.js\n\n// There is a default and a named export.\nexport const foo = 'foo';\nconst bar = 'bar';\nexport default 'bar';",
      errors: [message],
    },
  ],
});
