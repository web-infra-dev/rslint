import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-rest-params', {
  valid: [
    'arguments;',
    'function foo(arguments: any) { arguments; }',
    'function foo() { var arguments: any; arguments; }',
    'var foo = () => arguments;',
    'function foo(...args: any[]) { args; }',
    'function foo() { arguments.length; }',
    'function foo() { arguments.callee; }',
  ],
  invalid: [
    {
      code: 'function foo() { arguments; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    {
      code: 'function foo() { arguments[0]; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    {
      code: 'function foo() { arguments[1]; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
  ],
});
