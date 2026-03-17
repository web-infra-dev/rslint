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
    // var arguments in nested block — hoisted to function scope, shadows built-in
    'function foo() { if (true) { var arguments: any = []; } arguments; }',
    // let arguments in block — reference INSIDE the block refers to block-scoped variable
    'function foo() { { let arguments: any = 1; arguments; } }',
    // const arguments in block — same as above
    'function foo() { { const arguments: any = 1; arguments; } }',
    // for-of with let arguments — loop variable shadows inside loop body
    'function foo() { for (let arguments of []) { arguments; } }',
    // for-in with let arguments — same as above
    'function foo() { for (let arguments in {}) { arguments; } }',
    // catch clause parameter named arguments — shadows inside catch body
    'function foo() { try {} catch(arguments) { arguments; } }',
    // function expression named "arguments" — name does NOT shadow implicit arguments
    'var foo = function arguments() { };',
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
    // Computed member access with Symbol
    {
      code: 'function foo() { arguments[Symbol.iterator]; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Storing arguments in a variable
    {
      code: 'function foo() { var x = arguments; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // let arguments in a block does NOT shadow the implicit arguments outside the block
    {
      code: 'function foo() { { let arguments: any = 1; } arguments; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Arrow function — arguments refers to enclosing non-arrow function
    {
      code: 'function foo() { var f = () => arguments; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Function expression
    {
      code: 'var foo = function() { arguments; };',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Class constructor
    {
      code: 'class C { constructor() { arguments; } }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Destructuring parameter — does NOT shadow implicit arguments
    {
      code: 'function foo({ arguments: a }: any) { arguments; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Function expression named "arguments" — name does NOT shadow body arguments
    {
      code: 'var foo = function arguments() { arguments; };',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Catch parameter is block-scoped — does NOT shadow arguments outside catch
    {
      code: 'function foo() { try {} catch(arguments) {} arguments; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Shorthand property in object literal — IS a reference to arguments
    {
      code: 'function foo() { var x = { arguments }; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Method
    {
      code: 'var obj = { method() { arguments; } };',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Getter
    {
      code: 'var obj = { get x() { arguments; } };',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Setter
    {
      code: 'var obj = { set x(v: any) { arguments; } };',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Nested arrows — arguments passes through multiple arrow boundaries
    {
      code: 'function foo() { var f = () => () => arguments; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // typeof arguments
    {
      code: 'function foo() { typeof arguments; }',
      errors: [{ messageId: 'preferRestParams' }],
    },
    // Spread arguments
    {
      code: 'function foo() { bar(...arguments); }',
      errors: [{ messageId: 'preferRestParams' }],
    },
  ],
});
