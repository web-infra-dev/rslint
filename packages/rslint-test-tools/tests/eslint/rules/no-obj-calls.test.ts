import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-obj-calls', {
  valid: ['var x = Math.random();', 'JSON.parse(foo)'],
  invalid: [
    {
      code: 'Math();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'new Math();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
  ],
});
