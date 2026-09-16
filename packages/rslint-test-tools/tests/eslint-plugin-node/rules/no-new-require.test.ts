import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const noNewRequire = {
  messageId: 'noNewRequire',
  message: 'Unexpected use of new with require.',
  line: 1,
  column: 17,
  endLine: 1,
  endColumn: 42,
};

// Every case from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-new-require.js
// All documentation examples at the same tag are included below, with the
// duplicate incorrect example included once and rule-enabling comments omitted.
ruleTester.run(
  'no-new-require',
  {},
  {
    valid: [
      {
        name: 'require call',
        code: "var appHeader = require('app-header')",
      },
      {
        name: 'construct module export',
        code: "var AppHeader = new (require('app-header'))",
      },
      {
        name: 'construct exported property',
        code: "var AppHeader = new (require('headers').appHeader)",
      },
      {
        name: 'documentation: require call',
        code: "var appHeader = require('app-header');",
      },
      {
        name: 'documentation: construct module export',
        code: "var appHeader = new (require('app-header'));",
      },
      {
        name: 'documentation: separate require and construction',
        code: "var AppHeader = require('app-header');\nvar appHeader = new AppHeader();",
      },
    ],
    invalid: [
      {
        name: 'new require',
        code: "var appHeader = new require('app-header')",
        errors: [noNewRequire],
      },
      {
        name: 'property access after new require',
        code: "var appHeader = new require('headers').appHeader",
        errors: [{ ...noNewRequire, endColumn: 39 }],
      },
      {
        name: 'documentation: new require',
        code: "var appHeader = new require('app-header');",
        errors: [noNewRequire],
      },
    ],
  },
);
