import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-eq-null', {
  valid: ['if (x === null) { }', 'if (null === f()) { }'],
  invalid: [
    {
      code: 'if (x == null) { }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if (x != null) { }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'do {} while (null == x)',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
