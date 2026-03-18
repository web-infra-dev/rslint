import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-new-symbol', {
  valid: [
    // Basic non-constructor usage
    "var foo = Symbol('foo');",
    'new foo(Symbol);',
    'new foo(bar, Symbol);',

    // Shadowing by declaration type
    'var Symbol = function() {}; new Symbol();',
    'let Symbol = 1; new Symbol();',
    'const Symbol = null; new Symbol();',
    'function Symbol() {} new Symbol();',
    'class Symbol {} new Symbol();',

    // Shadowing by parameter type
    "function bar(Symbol) { var baz = new Symbol('baz'); }",
    'const f = (Symbol) => { new Symbol(); };',
    'function f(...Symbol) { new Symbol(); }',
    'function f({ Symbol }) { new Symbol(); }',
    'function f(Symbol = 1) { new Symbol(); }',

    // Shadowing by destructuring
    'var { Symbol } = obj; new Symbol();',
    'var [Symbol] = arr; new Symbol();',
    'var { a: { Symbol } } = obj; new Symbol();',
    'var { ["Symbol"]: Symbol } = obj; new Symbol();',

    // Shadowing by loop variable
    'for (var Symbol = 0;;) { new Symbol(); }',
    'for (var Symbol in obj) { new Symbol(); }',
    'for (let Symbol of arr) { new Symbol(); }',
    'for (const Symbol of arr) { new Symbol(); }',

    // Shadowing by catch clause
    'try {} catch(Symbol) { new Symbol(); }',

    // Top-level shadow affects all inner scopes
    'function Symbol() {} function f() { new Symbol(); }',
    'var Symbol = 1; function f() { new Symbol(); }',
    'var Symbol = 1; const f = () => new Symbol();',
    'var Symbol = 1; class C { m() { new Symbol(); } }',

    // Hoisting: var/function declarations hoist to top of scope
    'new Symbol(); var Symbol = 1;',
    'new Symbol(); function Symbol() {}',
  ],
  invalid: [
    // Basic cases
    {
      code: "var foo = new Symbol('foo');",
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: 'new Symbol()',
      errors: [{ messageId: 'noNewSymbol' }],
    },

    // Nested function expression doesn't shadow global
    {
      code: "function bar() { return function Symbol() {}; } var baz = new Symbol('baz');",
      errors: [{ messageId: 'noNewSymbol' }],
    },

    // Block-scoped declarations don't shadow outside the block
    {
      code: '{ function Symbol() {} } new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: '{ let Symbol = 1; } new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: '{ const Symbol = 1; } new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: 'if (true) { let Symbol = 1; } new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: 'try { let Symbol = 1; } catch(e) {} new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },

    // Function/arrow scoped var doesn't shadow outside
    {
      code: 'function foo() { var Symbol = 1; } new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: 'const f = () => { var Symbol = 1; }; new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: 'function a() { function b() { var Symbol = 1; } } new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },

    // IIFE parameters don't shadow outside
    {
      code: '(function(Symbol) {})(1); new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: '((Symbol) => {})(1); new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },

    // Class static block scoped var doesn't shadow outside
    {
      code: 'class C { static { var Symbol = 1; } } new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },

    // new Symbol() inside scopes without any shadow
    {
      code: 'function f() { new Symbol(); }',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: 'const f = () => new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: 'class C { m() { new Symbol(); } }',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: 'if (true) { if (true) { new Symbol(); } }',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: 'class C { static { new Symbol(); } }',
      errors: [{ messageId: 'noNewSymbol' }],
    },

    // Mixed: shadow in one scope, global in another
    {
      code: 'function f(Symbol) { new Symbol(); } new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
    {
      code: '{ let Symbol = 1; new Symbol(); } new Symbol();',
      errors: [{ messageId: 'noNewSymbol' }],
    },
  ],
});
