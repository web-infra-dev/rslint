import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-self-assign', {
  valid: ['var a = a', 'a = b', 'a += a', 'a = +a', 'a = [a]'],
  invalid: [
    {
      code: 'a = a',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a.b = a.b',
      errors: [{ messageId: 'selfAssignment' }],
    },
    {
      code: 'a.b.c = a.b.c',
      errors: [{ messageId: 'selfAssignment' }],
    },
  ],
});
