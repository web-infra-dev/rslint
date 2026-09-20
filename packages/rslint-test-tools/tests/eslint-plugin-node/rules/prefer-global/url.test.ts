import { RuleTester } from '../../rule-tester';

const urlError = (
  messageId: 'preferGlobal' | 'preferModule',
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
) => ({
  messageId,
  message:
    messageId === 'preferGlobal'
      ? "Unexpected use of 'require(\"url\").URL'. Use the global variable 'URL' instead."
      : "Unexpected use of the global variable 'URL'. Use 'require(\"url\").URL' instead.",
  line,
  column,
  endLine,
  endColumn,
});

const provideModuleMethods = ['require', 'process.getBuiltinModule'];

// Every case and JavaScript documentation example from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/url.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/url.md
new RuleTester({
  languageOptions: {
    sourceType: 'commonjs',
    globals: {
      URL: 'readonly',
      require: 'readonly',
      process: 'readonly',
      global: 'readonly',
    },
  },
}).run(
  'prefer-global/url',
  {},
  {
    valid: [
      { code: 'var b = new URL(s)' },
      { code: 'var b = new URL(s)', options: ['always'] },
      ...provideModuleMethods.flatMap((method) => [
        {
          code: `var { URL } = ${method}('url'); var b = new URL(s)`,
          options: ['never'],
        },
        {
          code: `var { URL } = ${method}('node:url'); var b = new URL(s)`,
          options: ['never'],
        },
      ]),
      // Documentation: always / never.
      { code: 'const u = new URL(s)' },
      {
        code: 'const { URL } = require("url")\nconst u = new URL(s)',
        options: ['never'],
      },
    ],
    invalid: [
      ...provideModuleMethods.flatMap((method) => [
        {
          code: `var { URL } = ${method}('url'); var b = new URL(s)`,
          errors: [urlError('preferGlobal', 1, 7, 1, 10)],
        },
        {
          code: `var { URL } = ${method}('node:url'); var b = new URL(s)`,
          errors: [urlError('preferGlobal', 1, 7, 1, 10)],
        },
        {
          code: `var { URL } = ${method}('url'); var b = new URL(s)`,
          options: ['always'],
          errors: [urlError('preferGlobal', 1, 7, 1, 10)],
        },
        {
          code: `var { URL } = ${method}('node:url'); var b = new URL(s)`,
          options: ['always'],
          errors: [urlError('preferGlobal', 1, 7, 1, 10)],
        },
      ]),
      {
        code: 'var b = new URL(s)',
        options: ['never'],
        errors: [urlError('preferModule', 1, 13, 1, 16)],
      },
      // Documentation: equivalence / always / never.
      {
        code: 'console.log(URL === require("url").URL) //→ true',
        errors: [urlError('preferGlobal', 1, 21, 1, 39)],
      },
      {
        code: 'const { URL } = require("url")\nconst u = new URL(s)',
        errors: [urlError('preferGlobal', 1, 9, 1, 12)],
      },
      {
        code: 'const u = new URL(s)',
        options: ['never'],
        errors: [urlError('preferModule', 1, 15, 1, 18)],
      },
    ],
  },
);
