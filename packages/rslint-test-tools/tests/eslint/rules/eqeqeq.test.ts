import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('eqeqeq', {
  valid: ['a === b', 'a !== b'],
  invalid: [
    {
      code: 'a == b',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'a != b',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
