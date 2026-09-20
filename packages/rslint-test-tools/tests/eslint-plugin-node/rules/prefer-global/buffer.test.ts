import { RuleTester } from '../../rule-tester';

const preferBufferAt = (
  messageId: 'preferGlobal' | 'preferModule',
  column: number,
  endColumn: number,
) => ({
  messageId,
  message:
    messageId === 'preferGlobal'
      ? `Unexpected use of 'require("buffer").Buffer'. Use the global variable 'Buffer' instead.`
      : `Unexpected use of the global variable 'Buffer'. Use 'require("buffer").Buffer' instead.`,
  line: 1,
  column,
  endLine: 1,
  endColumn,
});

const provideModuleMethods = ['require', 'process.getBuiltinModule'];

// Every upstream case and JavaScript documentation example at eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/buffer.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/buffer.md
new RuleTester({
  languageOptions: {
    sourceType: 'commonjs',
    globals: {
      Buffer: 'readonly',
      require: 'readonly',
      process: 'readonly',
      global: 'readonly',
    },
  },
}).run(
  'prefer-global/buffer',
  {},
  {
    valid: [
      { code: 'var b = Buffer.alloc(10)' },
      { code: 'var b = Buffer.alloc(10)', options: ['always'] },
      ...provideModuleMethods.flatMap((method) =>
        ['buffer', 'node:buffer'].map((module) => ({
          code: `var { Buffer } = ${method}('${module}'); var b = Buffer.alloc(10)`,
          options: ['never'],
        })),
      ),
      // Documentation: always and never.
      { code: 'const b = Buffer.alloc(16)' },
      {
        code: 'const { Buffer } = require("buffer")\nconst b = Buffer.alloc(16)',
        options: ['never'],
      },
    ],
    invalid: [
      ...provideModuleMethods.flatMap((method) =>
        ['buffer', 'node:buffer'].flatMap((module) =>
          [[], ['always']].map((options) => ({
            code: `var { Buffer } = ${method}('${module}'); var b = Buffer.alloc(10)`,
            options,
            errors: [preferBufferAt('preferGlobal', 7, 13)],
          })),
        ),
      ),
      {
        code: 'var b = Buffer.alloc(10)',
        options: ['never'],
        errors: [preferBufferAt('preferModule', 9, 15)],
      },
      // Documentation: identity comparison, always and never.
      {
        code: 'console.log(Buffer === require("buffer").Buffer) //→ true',
        errors: [preferBufferAt('preferGlobal', 24, 48)],
      },
      {
        code: 'const { Buffer } = require("buffer")\nconst b = Buffer.alloc(16)',
        errors: [preferBufferAt('preferGlobal', 9, 15)],
      },
      {
        code: 'const b = Buffer.alloc(16)',
        options: ['never'],
        errors: [preferBufferAt('preferModule', 11, 17)],
      },
    ],
  },
);
