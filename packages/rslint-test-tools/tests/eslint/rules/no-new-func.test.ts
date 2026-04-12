import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-new-func', {
  valid: [
    // --- Not the global Function ---
    'var a = new _function("b", "c", "return b+c");',
    'var a = _function("b", "c", "return b+c");',

    // --- Function as a value reference, not invoked ---
    'call(Function)',
    'new Class(Function)',
    'foo[Function]()',
    'var x = [Function]',
    'var x = Function',

    // --- Non-matching method calls ---
    'Function.toString()',
    'Function.hasOwnProperty("call")',

    // --- Dynamic/computed property: not statically "call"/"apply"/"bind" ---
    'Function[call]()',

    // --- Accessing but not calling .bind/.call/.apply ---
    'foo(Function.bind)',
    'var x = Function.call',

    // --- Shadowing: class declaration ---
    {
      code: 'class Function {}; new Function()',
    },
    {
      code: 'const fn = () => { class Function {}; new Function() }',
    },

    // --- Shadowing: function declaration ---
    'function Function() {}; Function()',
    'var fn = function () { function Function() {}; Function() }',

    // --- Shadowing: function expression name ---
    'var x = function Function() { Function(); }',

    // --- Shadowing: var (hoisted across blocks) ---
    'function test() { var Function = function(){}; return new Function(); }',
    "function test() { var x = new Function('code'); var Function = function() {}; }",
    'function test() { if (true) { var Function = 42; } new Function(); }',
    'function test() { for (var Function = 0; Function < 1; Function++) {} new Function(); }',
    'function test() { for (var Function in {}) {} new Function(); }',
    'function test() { for (var Function of []) {} new Function(); }',
    'function test() { switch (0) { case 0: var Function = 1; } new Function(); }',

    // --- Shadowing: let/const ---
    'function test() { let Function = class {}; return new Function(); }',
    'function test() { const Function = class {}; return Function(); }',

    // --- Shadowing: parameter ---
    'function test(Function) { return new Function(); }',
    'function test({ Function }) { return new Function(); }',
    'function test([Function]) { return new Function(); }',
    'function test(...Function) { return new Function(); }',
    'var fn = (Function) => Function();',
    'function* gen(Function) { yield new Function(); }',
    'async function af(Function) { return new Function(); }',

    // --- Shadowing: catch clause ---
    'try {} catch (Function) { new Function(); }',

    // --- Shadowing: nested scopes ---
    'function test() { var Function = class {}; function inner() { return new Function(); } }',
    'function test() { var Function = class {}; var fn = () => new Function(); }',

    // --- Shadowing: method/constructor parameters ---
    'var obj = { m(Function) { return new Function(); } };',
    'class C { m(Function) { return new Function(); } }',
    'class C { constructor(Function) { this.x = new Function(); } }',

    // --- Shadowing: for-let/of (inside loop body) ---
    'function test() { for (let Function in {}) { new Function(); } }',
    'function test() { for (let Function of []) { new Function(); } }',

    // --- Shadowing applies to .call/.apply/.bind too ---
    'function test(Function) { return Function.call(null, "code"); }',
    'function test() { var Function = class {}; Function.apply(null, ["code"]); }',

    // --- Tagged template (not a call) ---
    'Function`code`',
  ],
  invalid: [
    // === Direct: new Function(...) ===
    {
      code: 'var a = new Function("b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'new Function()',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Direct: Function(...) ===
    {
      code: 'var a = Function("b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Parenthesized callee ===
    {
      code: '(Function)("code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: '((Function))("code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'new (Function)("code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Optional call ===
    {
      code: 'Function?.("code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Method: .call / .apply / .bind (dot notation) ===
    {
      code: 'var a = Function.call(null, "b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function.apply(null, ["b", "c", "return b+c"]);',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function.bind(null, "b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    // .bind(...)() — only the inner Function.bind(...) is reported
    {
      code: 'var a = Function.bind(null, "b", "c", "return b+c")();',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Method: bracket notation ===
    {
      code: 'var a = Function["call"](null, "b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function["apply"](null, ["b", "c", "return b+c"]);',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var a = Function["bind"](null, "b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    // Template literal bracket notation
    {
      code: 'var a = Function[`call`](null, "code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Optional chaining on method ===
    {
      code: '(Function?.call)(null, "b", "c", "return b+c");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'Function?.call(null, "code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'Function?.apply(null, ["code"])',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'Function?.bind(null, "code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Parenthesized object in method call ===
    {
      code: '(Function).call(null, "code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: '(Function).apply(null, ["code"])',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === TypeScript assertions on callee ===
    {
      code: '(Function as any)("code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: '(<any>Function)("code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'Function!("code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: '(Function satisfies any)("code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'new (Function as any)("code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    // TypeScript assertion on object of method call
    {
      code: '(Function as any).call(null, "code")',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Nested new wrapping a Function call ===
    {
      code: 'new (Function("code"))',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Nesting: inside various constructs ===
    {
      code: 'function f() { return new Function("code"); }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var f = () => new Function("code");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'if (true) { var x = new Function("code"); }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'class C { m() { return new Function("code"); } }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'class C { constructor() { this.x = new Function("code"); } }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'function outer() { function inner() { return new Function("code"); } }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Multiple errors in one statement ===
    {
      code: 'var a = new Function("a"); var b = Function("b");',
      errors: [
        { messageId: 'noFunctionConstructor' },
        { messageId: 'noFunctionConstructor' },
      ],
    },

    // === Expressions: ternary, logical ===
    {
      code: 'var x = true ? new Function("a") : Function("b");',
      errors: [
        { messageId: 'noFunctionConstructor' },
        { messageId: 'noFunctionConstructor' },
      ],
    },
    {
      code: 'var x = foo || new Function("code");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: 'var x = foo ?? new Function("code");',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },

    // === Scoping: inner shadow does NOT reach outer ===
    {
      code: "const fn = () => { class Function {} }; new Function('', '')",
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    {
      code: "var fn = function () { function Function() {} }; Function('', '')",
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    // let in sibling block does not shadow
    {
      code: 'function f() { { let Function = class {}; } var x = new Function("code"); }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    // var in inner function does NOT hoist to outer
    {
      code: 'function f() { var x = new Function("code"); (function() { var Function = 1; })(); }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    // arrow param does not shadow outer scope
    {
      code: 'function f() { var fn = (Function) => Function; var x = new Function("code"); }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    // catch variable does not shadow outside catch block
    {
      code: 'function f() { try {} catch (Function) {} var x = new Function("code"); }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
    // for-let does not shadow outside the loop
    {
      code: 'function f() { for (let Function of []) {} var x = new Function("code"); }',
      errors: [{ messageId: 'noFunctionConstructor' }],
    },
  ],
});
