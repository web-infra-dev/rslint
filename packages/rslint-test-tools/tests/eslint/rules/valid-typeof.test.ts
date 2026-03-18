import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-typeof', {
  valid: [
    "typeof foo === 'string'",
    "typeof foo === 'number'",
    'typeof foo === typeof bar',
  ],
  invalid: [
    {
      code: "typeof foo === 'strnig'",
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: "typeof foo === 'nubmer'",
      errors: [{ messageId: 'invalidValue' }],
    },
  ],
});
