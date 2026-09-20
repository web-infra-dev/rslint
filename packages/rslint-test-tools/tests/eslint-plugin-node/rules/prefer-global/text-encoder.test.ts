import { RuleTester } from '../../rule-tester';

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/text-encoder.js
const ruleTester = new RuleTester({
  languageOptions: {
    globals: {
      TextEncoder: 'readonly',
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
        ? `Unexpected use of 'require("util").TextEncoder'. Use the global variable 'TextEncoder' instead.`
        : `Unexpected use of the global variable 'TextEncoder'. Use 'require("util").TextEncoder' instead.`,
    line,
    column,
    endLine,
    endColumn,
  };
}

const provideModuleMethods = ['require', 'process.getBuiltinModule'];
ruleTester.run(
  'prefer-global/text-encoder',
  {},
  {
    valid: [
      { code: 'var b = new TextEncoder(s)' },
      { code: 'var b = new TextEncoder(s)', options: ['always'] },
      ...provideModuleMethods.flatMap((method) =>
        ['util', 'node:util'].map((module) => ({
          code: `var { TextEncoder } = ${method}('${module}'); var b = new TextEncoder(s)`,
          options: ['never'],
        })),
      ),
      // Documentation examples (rule-enabling comments omitted).
      { code: 'const u = new TextEncoder(s)' },
      {
        code: 'const { TextEncoder } = require("util")\nconst u = new TextEncoder(s)',
        options: ['never'],
      },
    ],
    invalid: [
      ...provideModuleMethods.flatMap((method) =>
        ['util', 'node:util'].flatMap((module) =>
          [undefined, ['always']].map((options) => ({
            code: `var { TextEncoder } = ${method}('${module}'); var b = new TextEncoder(s)`,
            options,
            errors: [errorAt('preferGlobal', 1, 7, 1, 18)],
          })),
        ),
      ),
      {
        code: 'var b = new TextEncoder(s)',
        options: ['never'],
        errors: [errorAt('preferModule', 1, 13, 1, 24)],
      },
      // https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/text-encoder.md
      {
        code: 'console.log(TextEncoder === require("util").TextEncoder) //→ true',
        errors: [errorAt('preferGlobal', 1, 29, 1, 56)],
      },
      {
        code: 'const { TextEncoder } = require("util")\nconst u = new TextEncoder(s)',
        errors: [errorAt('preferGlobal', 1, 9, 1, 20)],
      },
      {
        code: 'const u = new TextEncoder(s)',
        options: ['never'],
        errors: [errorAt('preferModule', 1, 15, 1, 26)],
      },
    ],
  },
);
