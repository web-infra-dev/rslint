import { RuleTester } from '../../rule-tester';

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/text-decoder.js
const ruleTester = new RuleTester({
  languageOptions: {
    globals: {
      TextDecoder: 'readonly',
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
        ? `Unexpected use of 'require("util").TextDecoder'. Use the global variable 'TextDecoder' instead.`
        : `Unexpected use of the global variable 'TextDecoder'. Use 'require("util").TextDecoder' instead.`,
    line,
    column,
    endLine,
    endColumn,
  };
}

const provideModuleMethods = ['require', 'process.getBuiltinModule'];
ruleTester.run(
  'prefer-global/text-decoder',
  {},
  {
    valid: [
      { code: 'var b = new TextDecoder(s)' },
      { code: 'var b = new TextDecoder(s)', options: ['always'] },
      ...provideModuleMethods.flatMap((method) =>
        ['util', 'node:util'].map((module) => ({
          code: `var { TextDecoder } = ${method}('${module}'); var b = new TextDecoder(s)`,
          options: ['never'],
        })),
      ),
      // Documentation examples (rule-enabling comments omitted).
      { code: 'const u = new TextDecoder(s)' },
      {
        code: 'const { TextDecoder } = require("util")\nconst u = new TextDecoder(s)',
        options: ['never'],
      },
    ],
    invalid: [
      ...provideModuleMethods.flatMap((method) =>
        ['util', 'node:util'].flatMap((module) =>
          [undefined, ['always']].map((options) => ({
            code: `var { TextDecoder } = ${method}('${module}'); var b = new TextDecoder(s)`,
            options,
            errors: [errorAt('preferGlobal', 1, 7, 1, 18)],
          })),
        ),
      ),
      {
        code: 'var b = new TextDecoder(s)',
        options: ['never'],
        errors: [errorAt('preferModule', 1, 13, 1, 24)],
      },
      // https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/text-decoder.md
      {
        code: 'console.log(TextDecoder === require("util").TextDecoder) //→ true',
        errors: [errorAt('preferGlobal', 1, 29, 1, 56)],
      },
      {
        code: 'const { TextDecoder } = require("util")\nconst u = new TextDecoder(s)',
        errors: [errorAt('preferGlobal', 1, 9, 1, 20)],
      },
      {
        code: 'const u = new TextDecoder(s)',
        options: ['never'],
        errors: [errorAt('preferModule', 1, 15, 1, 26)],
      },
    ],
  },
);
