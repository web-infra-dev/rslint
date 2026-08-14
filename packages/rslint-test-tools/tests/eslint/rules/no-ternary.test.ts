import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-ternary', {
  valid: ['"x ? y";'],
  invalid: [
    {
      code: 'var foo = true ? thing : stuff;',
      errors: [{ messageId: 'noTernaryOperator', line: 1, column: 11 }],
    },
    {
      code: 'true ? thing() : stuff();',
      errors: [{ messageId: 'noTernaryOperator', line: 1, column: 1 }],
    },
    {
      code: 'function foo(bar) { return bar ? baz : qux; }',
      errors: [{ messageId: 'noTernaryOperator', line: 1, column: 28 }],
    },
  ],
});
