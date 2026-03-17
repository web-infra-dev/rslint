import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-func-assign', {
  valid: [
    'function foo() { var foo = 1; }',
    'function foo(foo: any) { foo = 1; }',
    'function foo() { var foo: any; foo = 1; }',
    'var foo = function() {}; foo = 1;',
  ],
  invalid: [
    {
      code: 'function foo() { foo = 1; }',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() {} foo = 1;',
      errors: [{ messageId: 'isAFunction' }],
    },
  ],
});
