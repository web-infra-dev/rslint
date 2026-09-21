import { RuleTester } from '../../rule-tester';

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/url-search-params.js
const ruleTester = new RuleTester({
  languageOptions: {
    globals: {
      URLSearchParams: 'readonly',
      require: 'readonly',
      process: 'readonly',
      global: 'readonly',
    },
  },
});

function errorAt(
  messageId: 'preferGlobal' | 'preferModule',
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
) {
  return {
    messageId,
    message:
      messageId === 'preferGlobal'
        ? `Unexpected use of 'require("url").URLSearchParams'. Use the global variable 'URLSearchParams' instead.`
        : `Unexpected use of the global variable 'URLSearchParams'. Use 'require("url").URLSearchParams' instead.`,
    line,
    column,
    endLine,
    endColumn,
  };
}

const provideModuleMethods = ['require', 'process.getBuiltinModule'];
ruleTester.run(
  'prefer-global/url-search-params',
  {},
  {
    valid: [
      { code: 'var b = new URLSearchParams(s)' },
      { code: 'var b = new URLSearchParams(s)', options: ['always'] },
      ...provideModuleMethods.flatMap((method) =>
        ['url', 'node:url'].map((module) => ({
          code: `var { URLSearchParams } = ${method}('${module}'); var b = new URLSearchParams(s)`,
          options: ['never'],
        })),
      ),
      // Documentation examples (rule-enabling comments omitted).
      { code: 'const u = new URLSearchParams(s)' },
      {
        code: 'const { URLSearchParams } = require("url")\nconst u = new URLSearchParams(s)',
        options: ['never'],
      },
    ],
    invalid: [
      ...provideModuleMethods.flatMap((method) =>
        ['url', 'node:url'].flatMap((module) =>
          [undefined, ['always']].map((options) => ({
            code: `var { URLSearchParams } = ${method}('${module}'); var b = new URLSearchParams(s)`,
            options,
            errors: [errorAt('preferGlobal', 1, 7, 1, 22)],
          })),
        ),
      ),
      {
        code: 'var b = new URLSearchParams(s)',
        options: ['never'],
        errors: [errorAt('preferModule', 1, 13, 1, 28)],
      },
      // https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/url-search-params.md
      {
        code: 'console.log(URLSearchParams === require("url").URLSearchParams) //→ true',
        errors: [errorAt('preferGlobal', 1, 33, 1, 63)],
      },
      {
        code: 'const { URLSearchParams } = require("url")\nconst u = new URLSearchParams(s)',
        errors: [errorAt('preferGlobal', 1, 9, 1, 24)],
      },
      {
        code: 'const u = new URLSearchParams(s)',
        options: ['never'],
        errors: [errorAt('preferModule', 1, 15, 1, 30)],
      },
    ],
  },
);
