import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unreachable', {
  valid: [
    'function foo() { return bar(); function bar() { return 1; } }',
    'function foo() { var x = 1; return x; }',
    // if without else: not fully terminal
    'function foo() { if (x) { return; } bar(); }',
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
    // Unreachable after if/else where both return
    {
      code: 'function foo() { if (x) { return 1; } else { return 2; } bar(); }',
      errors: [{ messageId: 'unreachableCode' }],
    },
    // Unreachable after try/catch where both return
    {
      code: 'function foo() { try { return 1; } catch(e) { return 2; } bar(); }',
      errors: [{ messageId: 'unreachableCode' }],
    },
  ],
});
