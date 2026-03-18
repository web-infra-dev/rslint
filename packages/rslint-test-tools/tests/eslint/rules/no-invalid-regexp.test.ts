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
    // Duplicate flags
    {
      code: "new RegExp('.', 'gg');",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Invalid pattern with non-literal flags — still reported
    {
      code: "new RegExp('[', flags);",
      errors: [{ messageId: 'regexMessage' }],
    },
  ],
});
