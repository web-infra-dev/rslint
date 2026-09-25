import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/exports-last.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/exports-last.md
const message = 'Export statements should appear at the end of the file';

const ruleTester = new RuleTester();
ruleTester.run('exports-last', null as never, {
  valid: [
    // Empty file.
    {
      code: '// comment',
    },
    // No exports.
    {
      code: "\n        const foo = 'bar'\n        const bar = 'baz'\n      ",
    },
    // Named export.
    {
      code: "\n        const foo = 'bar'\n        export {foo}\n      ",
    },
    // Default export.
    {
      code: "\n        const foo = 'bar'\n        export default foo\n      ",
    },
    // Only exports.
    {
      code: '\n        export default foo\n        export const bar = true\n      ',
    },
    // Statements before exports.
    {
      code: "\n        const foo = 'bar'\n        export default foo\n        export const bar = true\n      ",
    },
    // Multiline export.
    {
      code: "\n        const foo = 'bar'\n        export default function bar () {\n          const very = 'multiline'\n        }\n        export const baz = true\n      ",
    },
    // Many exports.
    {
      code: "\n        const foo = 'bar'\n        export default foo\n        export const so = 'many'\n        export const exports = ':)'\n        export const i = 'cant'\n        export const even = 'count'\n        export const how = 'many'\n      ",
    },
    // Export all.
    {
      code: "\n        export * from './foo'\n      ",
    },
    // Documentation example.
    {
      code: "const arr = ['bar']\n\nexport const bool = true\n\nexport default bool\n\nexport function func() {\n  console.log('Hello World 🌍')\n}\n\nexport const str = 'foo'\n",
    },
  ],
  invalid: [
    // Default export before a variable declaration.
    {
      code: "\n        export default 'bar'\n        const bar = true\n      ",
      errors: [message],
    },
    // Named export before a variable declaration.
    {
      code: "\n        export const foo = 'bar'\n        const bar = true\n      ",
      errors: [message],
    },
    // Export all before a variable declaration.
    {
      code: "\n        export * from './foo'\n        const bar = true\n      ",
      errors: [message],
    },
    // Many exports around a variable declaration.
    {
      code: "\n        export default 'such foo many bar'\n        export const so = 'many'\n        const foo = 'bar'\n        export const exports = ':)'\n        export const i = 'cant'\n        export const even = 'count'\n        export const how = 'many'\n      ",
      errors: [message, message],
    },
    // Documentation examples.
    {
      code: "\nconst bool = true\n\nexport default bool\n\nconst str = 'foo'\n\n",
      errors: [message],
    },
    {
      code: "\nexport const bool = true\n\nconst str = 'foo'\n\n",
      errors: [message],
    },
  ],
});
