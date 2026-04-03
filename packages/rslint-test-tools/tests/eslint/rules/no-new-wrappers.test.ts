import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-new-wrappers', {
  valid: [
    // Non-wrapper constructors
    'var a = new Object();',
    'var a = new Map();',
    'var a = new Date();',
    // Function call (not constructor)
    "var a = String('test'), b = String.fromCharCode(32);",
    // --- Shadowing: function parameters ---
    'function test(Number) { return new Number; }',
    'function test({ Number }) { return new Number(); }',
    'function test([Boolean]) { return new Boolean(); }',
    'function test({ a: { String } }) { return new String(); }',
    'function test(...Number) { return new Number(); }',
    'function test(Boolean = true) { return new Boolean(); }',
    'function test(a, Boolean, c) { return new Boolean(); }',
    'var fn = (String) => new String();',
    'function* gen(String) { yield new String(); }',
    'async function af(Number) { return new Number(); }',
    'var af = async (Boolean) => new Boolean();',
    // --- Shadowing: var (hoisted) ---
    'function test() { var Boolean = function(){}; return new Boolean(true); }',
    "function test() { var x = new String('hello'); var String = function() {}; }",
    'function test() { if (true) { var Number = 42; } var v = new Number(); }',
    'function test() { for (var Boolean = 0; Boolean < 1; Boolean++) {} new Boolean(); }',
    'function test() { for (var String in {}) {} new String(); }',
    'function test() { for (var Number of []) {} new Number(); }',
    'function test() { switch (0) { case 0: var Number = 1; } new Number(); }',
    // --- Shadowing: let/const (block-scoped) ---
    'function test() { let String = class {}; return new String("x"); }',
    'function test() { const Boolean = class {}; return new Boolean(); }',
    // --- Shadowing: function/class declaration ---
    'function test() { function Number() {} new Number(); }',
    'function test() { class String { constructor() {} }; new String(); }',
    // --- Shadowing: function expression name ---
    'var fn = function String() { return new String(); };',
    // --- Shadowing: catch clause ---
    'try {} catch (Number) { new Number(); }',
    // --- Shadowing: nested scopes ---
    'function test() { var Boolean = function() {}; function inner() { return new Boolean(); } }',
    'function test() { var String = class {}; function mid() { function deep() { return new String(); } } }',
    'function test() { var Number = class {}; var fn = () => new Number(); }',
    // --- Shadowing: method/constructor parameters ---
    'var obj = { m(Boolean) { return new Boolean(); } };',
    'class C { m(String) { return new String(); } }',
    'class C { constructor(Number) { this.x = new Number(); } }',
    // --- Shadowing: for-let/of (inside loop body) ---
    'function test() { for (let Boolean in {}) { new Boolean(); } }',
    'function test() { for (let String of []) { new String(); } }',
  ],
  invalid: [
    // --- Basic ---
    {
      code: "var a = new String('hello');",
      errors: [{ messageId: 'noConstructor' }],
    },
    {
      code: 'var a = new Number(10);',
      errors: [{ messageId: 'noConstructor' }],
    },
    {
      code: 'var a = new Boolean(false);',
      errors: [{ messageId: 'noConstructor' }],
    },
    // No parentheses
    {
      code: 'var a = new String;',
      errors: [{ messageId: 'noConstructor' }],
    },
    // Multiple in one statement
    {
      code: "var a = new String('a'); var b = new Number(1);",
      errors: [{ messageId: 'noConstructor' }, { messageId: 'noConstructor' }],
    },
    // --- Nesting: inside various constructs ---
    {
      code: "function f() { return new String('x'); }",
      errors: [{ messageId: 'noConstructor' }],
    },
    {
      code: 'var f = () => new Number(1);',
      errors: [{ messageId: 'noConstructor' }],
    },
    {
      code: 'if (true) { var x = new Boolean(false); }',
      errors: [{ messageId: 'noConstructor' }],
    },
    {
      code: 'class C { m() { return new Number(42); } }',
      errors: [{ messageId: 'noConstructor' }],
    },
    {
      code: 'class C { constructor() { this.x = new Boolean(true); } }',
      errors: [{ messageId: 'noConstructor' }],
    },
    {
      code: 'function outer() { function inner() { return new Number(1); } }',
      errors: [{ messageId: 'noConstructor' }],
    },
    // --- Scoping: shadow does NOT reach ---
    // let in sibling block
    {
      code: "function f() { { let String = class {}; } var x = new String('out'); }",
      errors: [{ messageId: 'noConstructor' }],
    },
    // var in inner function does NOT hoist to outer
    {
      code: 'function f() { var x = new Boolean(true); (function() { var Boolean = 1; })(); }',
      errors: [{ messageId: 'noConstructor' }],
    },
    // arrow param does not shadow outer scope
    {
      code: 'function f() { var fn = (Number) => Number; var x = new Number(1); }',
      errors: [{ messageId: 'noConstructor' }],
    },
    // catch variable does not shadow outside catch block
    {
      code: "function f() { try {} catch (String) {} var x = new String('after'); }",
      errors: [{ messageId: 'noConstructor' }],
    },
    // for-let does not shadow outside the loop
    {
      code: 'function f() { for (let Boolean of []) {} var x = new Boolean(false); }',
      errors: [{ messageId: 'noConstructor' }],
    },
    // --- Expressions ---
    {
      code: "var x = true ? new String('a') : new Number(1);",
      errors: [{ messageId: 'noConstructor' }, { messageId: 'noConstructor' }],
    },
  ],
});
