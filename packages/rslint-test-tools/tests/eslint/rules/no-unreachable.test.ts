import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unreachable', {
  valid: [
    'function foo() { return bar(); function bar() { return 1; } }',
    'function foo() { var x = 1; return x; }',
  ],
  invalid: [
    {
      code: 'function foo() { return; x = 1; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    {
      code: 'function foo() { throw error; x = 1; }',
      errors: [{ messageId: 'unreachableCode' }],
    },
  ],
});
