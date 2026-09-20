import { RuleTester } from '../../rule-tester';

const ruleTester = new RuleTester({
  languageOptions: {
    sourceType: 'commonjs',
    globals: {
      console: 'readonly',
      require: 'readonly',
      process: 'readonly',
    },
  },
});

const consoleError = (
  messageId: 'preferGlobal' | 'preferModule',
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
) => ({
  messageId,
  message:
    messageId === 'preferGlobal'
      ? "Unexpected use of 'require(\"console\")'. Use the global variable 'console' instead."
      : "Unexpected use of the global variable 'console'. Use 'require(\"console\")' instead.",
  line,
  column,
  endLine,
  endColumn,
});

const provideModuleMethods = ['require', 'process.getBuiltinModule'];

// Every case from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/console.js
ruleTester.run(
  'prefer-global/console',
  {},
  {
    valid: [
      { name: 'global console by default', code: 'console.log(10)' },
      {
        name: 'global console with always',
        code: 'console.log(10)',
        options: ['always'],
      },
      ...provideModuleMethods.flatMap((method) =>
        ['console', 'node:console'].map((module) => ({
          name: `${method}('${module}') with never`,
          code: `var console = ${method}('${module}'); console.log(10)`,
          options: ['never'],
        })),
      ),
    ],
    invalid: [
      ...provideModuleMethods.flatMap((method) =>
        ['console', 'node:console'].flatMap((module) =>
          [[], ['always']].map((options) => {
            const load = `${method}('${module}')`;
            return {
              name: `${load} with ${options[0] ?? 'default'}`,
              code: `var console = ${load}; console.log(10)`,
              options,
              errors: [
                consoleError('preferGlobal', 1, 15, 1, 15 + load.length),
              ],
            };
          }),
        ),
      ),
      {
        name: 'global console with never',
        code: 'console.log(10)',
        options: ['never'],
        errors: [consoleError('preferModule', 1, 1, 1, 8)],
      },
    ],
  },
);

// Every JavaScript example from the documentation at the same tag.
ruleTester.run(
  'prefer-global/console',
  {},
  {
    valid: [
      { name: 'documentation: global console', code: 'console.log("hello")' },
      {
        name: 'documentation: module console with never',
        code: 'const console = require("console")\nconsole.log("hello")',
        options: ['never'],
      },
    ],
    invalid: [
      {
        name: 'documentation: global and module identity',
        code: 'console.log(console === require("console")) //→ true',
        errors: [consoleError('preferGlobal', 1, 25, 1, 43)],
      },
      {
        name: 'documentation: module console by default',
        code: 'const console = require("console")\nconsole.log("hello")',
        errors: [consoleError('preferGlobal', 1, 17, 1, 35)],
      },
      {
        name: 'documentation: global console with never',
        code: 'console.log("hello")',
        options: ['never'],
        errors: [consoleError('preferModule', 1, 1, 1, 8)],
      },
    ],
  },
);
