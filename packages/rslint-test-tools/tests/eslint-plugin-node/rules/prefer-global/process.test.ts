import { RuleTester } from '../../rule-tester';

const ruleTester = new RuleTester({
  languageOptions: {
    sourceType: 'commonjs',
    globals: {
      process: 'readonly',
      require: 'readonly',
      global: 'readonly',
    },
  },
});
const preferGlobal = {
  messageId: 'preferGlobal',
  message:
    "Unexpected use of 'require(\"process\")'. Use the global variable 'process' instead.",
  line: 1,
  column: 16,
  endLine: 1,
  endColumn: 34,
};
const preferModule = {
  messageId: 'preferModule',
  message:
    "Unexpected use of the global variable 'process'. Use 'require(\"process\")' instead.",
  line: 1,
  column: 1,
  endLine: 1,
  endColumn: 8,
};

// Every upstream test and JavaScript documentation example at eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-global/process.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-global/process.md
ruleTester.run(
  'prefer-global/process',
  {},
  {
    valid: [
      { name: 'default global usage', code: 'process.exit(0)' },
      {
        name: 'explicit global usage',
        code: 'process.exit(0)',
        options: ['always'],
      },
      {
        name: 'module usage',
        code: "var process = require('process'); process.exit(0)",
        options: ['never'],
      },
      {
        name: 'module usage with node protocol',
        code: "var process = require('node:process'); process.exit(0)",
        options: ['never'],
      },
      {
        name: 'unrelated builtin module',
        code: "process.getBuiltinModule('buffer')",
        options: ['always'],
      },
      { name: 'documentation: always', code: 'process.exit(0)' },
      {
        name: 'documentation: never',
        code: 'const process = require("process")\nprocess.exit(0)',
        options: ['never'],
      },
    ],
    invalid: [
      ...['require', 'process.getBuiltinModule'].flatMap((method) =>
        [undefined, ['always']].flatMap((options) =>
          ['process', 'node:process'].map((module) => {
            const call = `${method}('${module}')`;
            return {
              name: `${method}('${module}') with ${options?.[0] ?? 'default'} option`,
              code: `var process_ = ${call}; process_.exit(0)`,
              options,
              errors: [{ ...preferGlobal, endColumn: 16 + call.length }],
            };
          }),
        ),
      ),
      {
        name: 'global usage with never option',
        code: 'process.exit(0)',
        options: ['never'],
        errors: [preferModule],
      },
      {
        name: 'documentation: overview',
        code: 'process.log(process === require("process"))',
        errors: [{ ...preferGlobal, column: 25, endColumn: 43 }],
      },
      {
        name: 'documentation: always',
        code: 'const process = require("process")\nprocess.exit(0)',
        errors: [{ ...preferGlobal, column: 17, endColumn: 35 }],
      },
      {
        name: 'documentation: never',
        code: 'process.exit(0)',
        options: ['never'],
        errors: [preferModule],
      },
    ],
  },
);
