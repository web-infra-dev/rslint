import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-invalid-regexp', {
  valid: [
    "RegExp('.')",
    "new RegExp('.')",
    "new RegExp('.', 'im')",
    "new RegExp(pattern, 'g')",
  ],
  invalid: [
    {
      code: "RegExp('.', 'z');",
      errors: [{ messageId: 'regexMessage' }],
    },
  ],
});
