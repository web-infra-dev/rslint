import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const noProcessExit = {
  messageId: 'noProcessExit',
  message: "Don't use process.exit(); throw an error instead.",
  line: 1,
  column: 1,
  endLine: 1,
  endColumn: 16,
};

// Every case from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-process-exit.js
// Documentation examples at the same tag are included below.
ruleTester.run(
  'no-process-exit',
  {},
  {
    valid: [
      { name: 'different object', code: 'Process.exit()' },
      { name: 'read exit property', code: 'var exit = process.exit;' },
      { name: 'pass exit as an argument', code: 'f(process.exit)' },
      {
        name: 'documentation: throw an error',
        code: `if (somethingBadHappened) {
    throw new Error("Something bad happened!");
}`,
      },
      {
        name: 'documentation: correct',
        code: 'Process.exit();\nvar exit = process.exit;',
      },
    ],
    invalid: [
      {
        name: 'exit with zero',
        code: 'process.exit(0);',
        errors: [noProcessExit],
      },
      {
        name: 'exit with one',
        code: 'process.exit(1);',
        errors: [noProcessExit],
      },
      {
        name: 'nested call',
        code: 'f(process.exit(1));',
        errors: [{ ...noProcessExit, column: 3, endColumn: 18 }],
      },
      {
        name: 'documentation: conditional exit',
        code: `if (somethingBadHappened) {
    console.error("Something bad happened!");
    process.exit(1);
}`,
        errors: [
          { ...noProcessExit, line: 3, column: 5, endLine: 3, endColumn: 20 },
        ],
      },
      {
        name: 'documentation: incorrect',
        code: 'process.exit(1);\nprocess.exit(0);',
        errors: [noProcessExit, { ...noProcessExit, line: 2, endLine: 2 }],
      },
    ],
  },
);
