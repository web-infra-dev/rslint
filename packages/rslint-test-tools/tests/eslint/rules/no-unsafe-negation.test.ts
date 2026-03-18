import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unsafe-negation', {
  valid: ['a in b', '!(a in b)', '(!a) in b'],
  invalid: [
    {
      code: '!a in b',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '!a instanceof b',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
