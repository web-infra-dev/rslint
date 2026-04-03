import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-caller', {
  valid: [
    // Basic valid cases
    'var x = arguments.length',
    'var x = arguments',
    'var x = arguments[0]',
    'var x = arguments[caller]',
    // Computed access with string literals (not PropertyAccessExpression)
    'var x = arguments["callee"]',
    'var x = arguments["caller"]',
    // callee/caller on non-arguments objects
    'var obj = { callee: 1 }; var x = obj.callee',
    'var obj = { caller: 1 }; var x = obj.caller',
    // Similar but non-matching property names
    'var x = arguments.called',
    'var x = arguments.call',
    'var x = arguments.callees',
    'var x = arguments.callers',
    // Comma operator result - not an arguments identifier
    'var x = (0, arguments).callee',
    // Ternary result - not an arguments identifier
    'var x = (true ? arguments : null).callee',
    // TypeScript non-null assertion wraps identifier (not flagged)
    'function nonNull() { arguments!.callee; }',
    // TypeScript as assertion wraps identifier (not flagged)
    'function typeAssert() { (arguments as any).callee; }',
    // TypeScript angle bracket assertion wraps identifier (not flagged)
    'function angleBracket() { (<any>arguments).callee; }',
  ],
  invalid: [
    // Basic cases
    {
      code: 'var x = arguments.callee',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var x = arguments.caller',
      errors: [{ messageId: 'unexpected' }],
    },
    // Chained property access
    {
      code: 'var x = arguments.callee.bind(this)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var x = arguments.caller.toString()',
      errors: [{ messageId: 'unexpected' }],
    },
    // In function call
    {
      code: 'function foo() { arguments.callee(); }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In function expression
    {
      code: 'var bar = function() { return arguments.callee; };',
      errors: [{ messageId: 'unexpected' }],
    },
    // In IIFE
    {
      code: '(function() { arguments.callee; })();',
      errors: [{ messageId: 'unexpected' }],
    },
    // Nested function - inner scope
    {
      code: 'function outer() {\n  function inner() {\n    arguments.callee;\n  }\n}',
      errors: [{ messageId: 'unexpected' }],
    },
    // In conditional
    {
      code: 'function cond() { if (true) { arguments.callee; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In loop
    {
      code: 'function loop() { for (var i = 0; i < 10; i++) { arguments.caller; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In ternary
    {
      code: 'function tern() { var x = true ? arguments.callee : null; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In logical expression
    {
      code: 'function logic() { var x = arguments.callee || null; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In assignment
    {
      code: 'function assign() { var x; x = arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In return statement
    {
      code: 'function ret() { return arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In throw statement
    {
      code: 'function thrw() { throw arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // As function argument
    {
      code: 'function arg() { console.log(arguments.callee); }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In arrow function (arguments from outer scope)
    {
      code: 'function arrowOuter() { var fn = () => arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In object literal value
    {
      code: 'function objLit() { var o = { fn: arguments.callee }; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In array literal
    {
      code: 'function arrLit() { var a = [arguments.callee]; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In template literal expression
    {
      code: 'function tmpl() { var s = `${arguments.callee}`; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In typeof
    {
      code: 'function typ() { var t = typeof arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In void
    {
      code: 'function vd() { void arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In delete
    {
      code: 'function del() { delete arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In comma expression
    {
      code: 'function comma() { (0, arguments.callee); }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In switch case
    {
      code: 'function sw() { switch(arguments.callee) { default: break; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In while condition
    {
      code: 'function wh() { while(arguments.callee) { break; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Multiple occurrences - callee calling caller
    {
      code: 'function nested() { arguments.callee(arguments.caller); }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    // In class method
    {
      code: 'class MyClass { method() { arguments.callee; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In try-catch
    {
      code: 'function tryCatch() { try { arguments.callee; } catch(e) {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Multiple on same line
    {
      code: 'function multi() { arguments.callee; arguments.caller; }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    // Parenthesized arguments - ESLint sees through parens
    {
      code: 'var x = (arguments).callee',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var x = ((arguments)).callee',
      errors: [{ messageId: 'unexpected' }],
    },
    // arguments as parameter name - still flagged (syntactic check)
    {
      code: 'function paramName(arguments) { arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In generator function
    {
      code: 'function* gen() { arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In async function
    {
      code: 'async function asyncFn() { arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In getter
    {
      code: 'var obj = { get x() { return arguments.callee; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In setter
    {
      code: 'var obj = { set x(v) { arguments.caller; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In method shorthand
    {
      code: 'var obj = { m() { arguments.callee; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Deeply nested control flow
    {
      code: 'function deep() {\n  if (true) {\n    while (true) {\n      for (var i = 0; i < 1; i++) {\n        try {\n          arguments.callee;\n        } catch(e) {}\n      }\n    }\n  }\n}',
      errors: [{ messageId: 'unexpected' }],
    },
    // Nested object value
    {
      code: 'function nestedObj() { var o = { a: { b: { c: arguments.callee } } }; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In tagged template
    {
      code: 'function tagged() { String.raw`${arguments.callee}`; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // With new expression
    {
      code: 'function newExpr() { new (arguments.callee)(); }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In spread
    {
      code: 'function spread() { var a = [...arguments.callee]; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // As computed property key
    {
      code: 'function compKey() { var o = { [arguments.callee]: 1 }; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // In sequence with assignment
    {
      code: 'function seq() { var x; x = (arguments.callee, 1); }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Optional chaining - still PropertyAccessExpression in TS AST
    {
      code: 'function optChain() { arguments?.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function optChain2() { arguments?.caller; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // var shadowing arguments - still flagged (syntactic check)
    {
      code: 'function varShadow() { var arguments = {}; arguments.callee; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // arguments in catch clause - still flagged
    {
      code: 'function catchArgs() { try {} catch(arguments) { arguments.callee; } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Top-level arguments.callee
    {
      code: 'arguments.callee',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'arguments.caller',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
