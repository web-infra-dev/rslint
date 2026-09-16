import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const noProcessEnvAt = (
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
) => ({
  messageId: 'unexpectedProcessEnv',
  message: 'Unexpected use of process.env.',
  line,
  column,
  endLine,
  endColumn,
});

// Every case from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-process-env.js
// Documentation examples at the same tag are included below.
ruleTester.run(
  'no-process-env',
  {},
  {
    valid: [
      {
        name: 'upstream valid 1',
        code: 'Process.env',
      },
      {
        name: 'upstream valid 2',
        code: 'process[env]',
      },
      {
        name: 'upstream valid 3',
        code: 'process.nextTick',
      },
      {
        name: 'upstream valid 4',
        code: 'process.execArgv',
      },
      // allowedVariables
      {
        name: 'upstream valid 5',
        code: 'process.env.NODE_ENV',
        options: [{ allowedVariables: ['NODE_ENV'] }],
      },
      {
        name: 'upstream valid 6',
        code: "process.env['NODE_ENV']",
        options: [{ allowedVariables: ['NODE_ENV'] }],
      },
      {
        name: 'upstream valid 7',
        code: "process['env'].NODE_ENV",
        options: [{ allowedVariables: ['NODE_ENV'] }],
      },
      {
        name: 'upstream valid 8',
        code: "process['env']['NODE_ENV']",
        options: [{ allowedVariables: ['NODE_ENV'] }],
      },
      {
        name: 'Use a configuration module.',
        code: 'var config = require("./config");\n\nif(config.env === "development") {\n    //...\n}',
      },
    ],
    invalid: [
      {
        name: 'upstream invalid 1',
        code: 'process.env',
        errors: [noProcessEnvAt(1, 1, 1, 12)],
      },
      {
        name: 'upstream invalid 2',
        code: "process['env']",
        errors: [noProcessEnvAt(1, 1, 1, 15)],
      },
      {
        name: 'upstream invalid 3',
        code: 'process.env.ENV',
        errors: [noProcessEnvAt(1, 1, 1, 12)],
      },
      {
        name: 'upstream invalid 4',
        code: 'f(process.env)',
        errors: [noProcessEnvAt(1, 3, 1, 14)],
      },
      // allowedVariables
      {
        name: 'upstream invalid 5',
        code: "process.env['OTHER_VARIABLE']",
        options: [{ allowedVariables: ['NODE_ENV'] }],
        errors: [noProcessEnvAt(1, 1, 1, 12)],
      },
      {
        name: 'upstream invalid 6',
        code: 'process.env.OTHER_VARIABLE',
        options: [{ allowedVariables: ['NODE_ENV'] }],
        errors: [noProcessEnvAt(1, 1, 1, 12)],
      },
      {
        name: 'upstream invalid 7',
        code: "process['env']['OTHER_VARIABLE']",
        options: [{ allowedVariables: ['NODE_ENV'] }],
        errors: [noProcessEnvAt(1, 1, 1, 15)],
      },
      {
        name: 'upstream invalid 8',
        code: "process['env'].OTHER_VARIABLE",
        options: [{ allowedVariables: ['NODE_ENV'] }],
        errors: [noProcessEnvAt(1, 1, 1, 15)],
      },
      {
        name: 'upstream invalid 9',
        code: 'process.env[NODE_ENV]',
        options: [{ allowedVariables: ['NODE_ENV'] }],
        errors: [noProcessEnvAt(1, 1, 1, 12)],
      },
      {
        name: 'upstream invalid 10',
        code: "process['env'][NODE_ENV]",
        options: [{ allowedVariables: ['NODE_ENV'] }],
        errors: [noProcessEnvAt(1, 1, 1, 15)],
      },
      {
        name: 'Read an environment variable directly.',
        code: 'if(process.env.NODE_ENV === "development") {\n    //...\n}',
        errors: [noProcessEnvAt(1, 4, 1, 15)],
      },
    ],
  },
);
