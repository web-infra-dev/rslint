// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('exports-last', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: '// comment',
    },
    {
      name: 'run 0 valid 1',
      code: "\n        const foo = 'bar'\n        const bar = 'baz'\n      ",
    },
    {
      name: 'run 0 valid 2',
      code: "\n        const foo = 'bar'\n        export {foo}\n      ",
    },
    {
      name: 'run 0 valid 3',
      code: "\n        const foo = 'bar'\n        export default foo\n      ",
    },
    {
      name: 'run 0 valid 4',
      code: '\n        export default foo\n        export const bar = true\n      ',
    },
    {
      name: 'run 0 valid 5',
      code: "\n        const foo = 'bar'\n        export default foo\n        export const bar = true\n      ",
    },
    {
      name: 'run 0 valid 6',
      code: "\n        const foo = 'bar'\n        export default function bar () {\n          const very = 'multiline'\n        }\n        export const baz = true\n      ",
    },
    {
      name: 'run 0 valid 7',
      code: "\n        const foo = 'bar'\n        export default foo\n        export const so = 'many'\n        export const exports = ':)'\n        export const i = 'cant'\n        export const even = 'count'\n        export const how = 'many'\n      ",
    },
    {
      name: 'run 0 valid 8',
      code: "\n        export * from './foo'\n      ",
    },
    {
      name: 'run 0 valid 9',
      code: "const arr = ['bar']\n\nexport const bool = true\n\nexport default bool\n\nexport function func() {\n  console.log('Hello World 🌍')\n}\n\nexport const str = 'foo'\n",
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "\n        export default 'bar'\n        const bar = true\n      ",
      errors: [
        {
          messageId: '',
          message: 'Export statements should appear at the end of the file',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "\n        export const foo = 'bar'\n        const bar = true\n      ",
      errors: [
        {
          messageId: '',
          message: 'Export statements should appear at the end of the file',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "\n        export * from './foo'\n        const bar = true\n      ",
      errors: [
        {
          messageId: '',
          message: 'Export statements should appear at the end of the file',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 30,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: "\n        export default 'such foo many bar'\n        export const so = 'many'\n        const foo = 'bar'\n        export const exports = ':)'\n        export const i = 'cant'\n        export const even = 'count'\n        export const how = 'many'\n      ",
      errors: [
        {
          messageId: '',
          message: 'Export statements should appear at the end of the file',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 43,
        },
        {
          messageId: '',
          message: 'Export statements should appear at the end of the file',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "\nconst bool = true\n\nexport default bool\n\nconst str = 'foo'\n\n",
      errors: [
        {
          messageId: '',
          message: 'Export statements should appear at the end of the file',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: "\nexport const bool = true\n\nconst str = 'foo'\n\n",
      errors: [
        {
          messageId: '',
          message: 'Export statements should appear at the end of the file',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 25,
        },
      ],
    },
  ],
});
