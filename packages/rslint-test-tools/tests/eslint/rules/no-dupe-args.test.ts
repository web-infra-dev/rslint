import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-dupe-args', {
  valid: [
    'function foo(a, b, c) {}',
    'var foo = function(a, b, c) {}',
    'function foo(a, b, c, d) {}',
  ],
  invalid: [
    {
      code: 'function foo(a, b, a) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    // Triple duplicate
    {
      code: 'function foo(a, a, a) {}',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: 'var foo = function(a, b, b) {}',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
