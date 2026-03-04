import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-dupe-args', {
  valid: [
    'function foo(a, b, c) {}',
    'var foo = function(a, b, c) {}',
    'function foo(a, b, c, d) {}',
    // Destructured params with different names
    'function foo(a, {b, c}) {}',
    'function foo({a}, {b}) {}',
    'function foo([a, b], c) {}',
  ],
  invalid: [
    {
      code: 'function foo(a, b, a) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    // Reports each duplicate occurrence (ESLint reports once per name)
    {
      code: 'function foo(a, a, a) {}',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: 'var foo = function(a, b, b) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    // Destructured parameter duplicate
    {
      code: 'function foo(a, {a}) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    // Array destructured parameter duplicate
    {
      code: 'function foo(a, [a]) {}',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
