// Every case from ESLint v10.8.1, with messages and full ranges
// additionally pinned from the same upstream version. No upstream expectations are relaxed.
// https://github.com/eslint/eslint/blob/v10.8.1/tests/lib/rules/no-debugger.js
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '',
});

ruleTester.run('no-debugger', {
  valid: [
    {
      code: 'var test = { debugger: 1 }; test.debugger;',
    },
  ],
  invalid: [
    {
      code: 'if (foo) debugger',
      errors: [
        {
          messageId: 'unexpected',
          message: "Unexpected 'debugger' statement.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
  ],
});
