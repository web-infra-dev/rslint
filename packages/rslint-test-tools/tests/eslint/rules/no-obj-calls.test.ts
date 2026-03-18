import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-obj-calls', {
  valid: [
    'var x = Math.random();',
    'JSON.parse(foo)',
    // Shadowed variable should not be flagged
    'function f() { var Math = 1; Math(); }',
  ],
  invalid: [
    {
      code: 'Math();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'new Math();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'JSON();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'Reflect();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'Atomics();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'Intl();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
  ],
});
