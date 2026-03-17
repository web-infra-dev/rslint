import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-inner-declarations', {
  valid: [
    'function doSomething() { }',
    'function doSomething() { function somethingElse() { } }',
    '(function() { function doSomething() { } }());',
    'if (test) { var fn = function() { }; }',
  ],
  invalid: [
    {
      code: 'if (foo) function f(){}',
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'function bar() { if (foo) function f(){}; }',
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
  ],
});
