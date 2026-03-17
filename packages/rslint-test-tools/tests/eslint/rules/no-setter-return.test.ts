import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-setter-return', {
  valid: [
    'var foo = { set a(val) { val = 1; } };',
    'class A { set a(val) { return; } }',
  ],
  invalid: [
    {
      code: 'var foo = { set a(val) { return 1; } };',
      errors: [{ messageId: 'setter' }],
    },
    {
      code: 'class A { set a(val) { return val; } }',
      errors: [{ messageId: 'setter' }],
    },
  ],
});
