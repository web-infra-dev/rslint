import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-eval', {
  valid: [
    // ================================================================
    // Basic: not eval
    // ================================================================
    'Eval(foo)',
    "setTimeout('foo')",
    "setInterval('foo')",
    "window.noeval('foo')",
    // global not declared in test env (no @types/node) → treated as unknown
    "global.eval('foo')",
    "global.noeval('foo')",
    "globalThis.noeval('foo')",
    "this.noeval('foo');",

    // ================================================================
    // Property / member names — NOT references to global eval
    // ================================================================
    'var obj = { eval: 42 }',
    'var obj = { eval: function() {} }',
    'var obj = { eval() { return 1; } }',
    'class A { eval = 42 }',
    'class A { eval() { return 1; } }',
    'class A { get eval() { return 1; } }',
    'class A { set eval(v) {} }',
    'class A { static eval() {} }',
    'class A { static eval = 42 }',
    'var { eval: e } = obj',
    'function foo({ eval: e }) { e() }',
    'interface I { eval: string }',
    'type T = { eval: string }',
    'enum E { eval }',
    'eval: while(true) { break eval; }',
    "import { eval as foo } from 'mod'; foo()",
    'var foo = 1; export { foo as eval }',
    // Re-export: eval is source module name, not local ref
    "export { eval } from 'mod'",
    "export { eval as foo } from 'mod'",
    "export { foo as eval } from 'mod'",

    // ================================================================
    // this.eval — safe contexts (this is NOT global)
    // ================================================================
    "function foo() { 'use strict'; this.eval('foo'); }",
    "'use strict'; function foo() { this.eval('foo'); }",
    "import x from 'y'; this.eval('foo');",
    "import x from 'y'; function foo() { this.eval('foo'); }",
    "export {}; () => { this.eval('foo') }",
    "var obj = {foo: function() { this.eval('foo'); }}",
    "var obj = {}; obj.foo = function() { this.eval('foo'); }",
    'var obj = { get foo() { return this.eval(); } }',
    'var obj = { set foo(v) { this.eval(v); } }',
    "function f() { 'use strict'; () => { this.eval('foo') } }",
    "(function f() { 'use strict'; () => { this.eval('foo') } })",
    "function f() { 'use strict'; () => () => this.eval('foo') }",
    'class A { foo() { this.eval(); } }',
    'class A { static foo() { this.eval(); } }',
    'class A { constructor() { this.eval(); } }',
    'class A { foo() { () => this.eval(); } }',
    'class A { foo() { () => () => this.eval(); } }',
    'class A { field = this.eval(); }',
    'class A { field = () => this.eval(); }',
    'class A { static { this.eval(); } }',
    "class C extends function () { this.eval('foo'); } {}",

    // ================================================================
    // Shadowing — eval is a local variable, not the global
    // ================================================================
    "function foo() { var eval = 'foo'; window[eval]('foo') }",
    'function foo(eval) { var x = eval }',
    'function foo() { let eval = 1; eval }',
    'function foo() { const eval = 1; eval }',
    'try {} catch(eval) { var x = eval }',
    'function eval() {} var x = eval',
    'var { a: eval } = obj; var x = eval',
    "import { eval } from 'mod'; var x = eval",

    // ================================================================
    // Uppercase constructor convention
    // ================================================================
    "var Foo = function() { this.eval('foo') }",
    "var MyClass = function() { this.eval('bar') }",

    // ================================================================
    // isDefaultThisBinding — not default
    // ================================================================
    "(function() { this.eval('foo') })()",
    "(function() { this.eval('foo') }).call(obj)",
    "(function() { this.eval('foo') }).apply(obj, [])",
    "var f = function() { this.eval('foo') }.bind(obj); f()",
    "obj.method = true ? function() { this.eval('foo') } : null",
    "obj.method = null || function() { this.eval('foo') }",
    "obj.foo = (function() { return function() { this.eval('foo') } })()",
    // Callback with thisArg
    "arr.forEach(function() { this.eval('foo') }, obj)",
    'arr.map(function(x) { return this.eval(x) }, obj)',
    'arr.filter(function(x) { return this.eval(x) }, obj)',
    'arr.find(function(x) { return this.eval(x) }, obj)',
    'arr.findIndex(function(x) { return this.eval(x) }, obj)',
    'arr.findLast(function(x) { return this.eval(x) }, obj)',
    'arr.findLastIndex(function(x) { return this.eval(x) }, obj)',
    'arr.flatMap(function(x) { return this.eval(x) }, obj)',
    'arr.some(function(x) { return this.eval(x) }, obj)',
    'arr.every(function(x) { return this.eval(x) }, obj)',
    "Reflect.apply(function() { this.eval('foo') }, obj, [])",
    'Array.from(iter, function(x) { return this.eval(x) }, obj)',
    // TS type assertion wrappers with IIFE
    "(function() { this.eval('foo') } as any)()",
    "(<any>function() { this.eval('foo') })()",

    // ================================================================
    // allowIndirect: true
    // ================================================================
    { code: "(0, eval)('foo')", options: [{ allowIndirect: true }] as any },
    {
      code: "(0, window.eval)('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "(0, window['eval'])('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "var EVAL = eval; EVAL('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "var EVAL = this.eval; EVAL('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "(function(exe){ exe('foo') })(eval);",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "window.eval('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "window.window.eval('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "window.window['eval']('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "global.eval('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "global.global.eval('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "this.eval('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "function foo() { this.eval('foo') }",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "(0, globalThis.eval)('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "(0, globalThis['eval'])('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "var EVAL = globalThis.eval; EVAL('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "function foo() { globalThis.eval('foo') }",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "globalThis.globalThis.eval('foo');",
      options: [{ allowIndirect: true }] as any,
    },
    { code: "eval?.('foo')", options: [{ allowIndirect: true }] as any },
    {
      code: "window?.eval('foo')",
      options: [{ allowIndirect: true }] as any,
    },
    {
      code: "(window?.eval)('foo')",
      options: [{ allowIndirect: true }] as any,
    },
  ],
  invalid: [
    // ================================================================
    // Direct eval calls
    // ================================================================
    { code: 'eval(foo)', errors: [{ messageId: 'unexpected' }] },
    { code: "eval('foo')", errors: [{ messageId: 'unexpected' }] },
    {
      code: "function foo(eval) { eval('foo') }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "function foo() { function bar() { eval('x') } }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "var fn = () => eval('foo')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "(function() { eval('foo') })()",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "async function foo() { eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "function* gen() { eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { foo() { eval('bar') } }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { field = eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { static { eval('bar') } }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "function foo(x = eval('bar')) {}",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Direct eval with allowIndirect: true
    // ================================================================
    {
      code: 'eval(foo)',
      options: [{ allowIndirect: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "eval('foo')",
      options: [{ allowIndirect: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "function foo(eval) { eval('foo') }",
      options: [{ allowIndirect: true }] as any,
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Indirect eval — flagged in default mode
    // ================================================================
    { code: "(0, eval)('foo')", errors: [{ messageId: 'unexpected' }] },

    // ================================================================
    // Global object access: globalThis (declared via esnext lib)
    // Note: window tests omitted here — JS test env has no DOM lib.
    // Window tests are covered in Go tests (which use a tsconfig with DOM).
    // ================================================================
    { code: "globalThis.eval('foo')", errors: [{ messageId: 'unexpected' }] },
    {
      code: "globalThis.globalThis.eval('foo')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "globalThis.globalThis['eval']('foo')",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Indirect eval via global object
    // ================================================================
    {
      code: "(0, globalThis.eval)('foo')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "(0, globalThis['eval'])('foo')",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Non-call eval references
    // ================================================================
    {
      code: "var EVAL = eval; EVAL('foo')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "(function(exe){ exe('foo') })(eval);",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "var EVAL = globalThis.eval; EVAL('foo')",
      errors: [{ messageId: 'unexpected' }],
    },
    { code: 'var arr = [eval]', errors: [{ messageId: 'unexpected' }] },
    {
      code: 'var fn = eval || function() {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var fn = true ? eval : null',
      errors: [{ messageId: 'unexpected' }],
    },
    { code: 'var obj = { eval }', errors: [{ messageId: 'unexpected' }] },
    { code: 'var arr = [...eval]', errors: [{ messageId: 'unexpected' }] },
    { code: 'var s = `${eval}`', errors: [{ messageId: 'unexpected' }] },
    { code: 'var obj = { [eval]: 42 }', errors: [{ messageId: 'unexpected' }] },
    { code: 'eval`template`', errors: [{ messageId: 'unexpected' }] },
    { code: 'typeof eval', errors: [{ messageId: 'unexpected' }] },
    { code: 'eval = function() {}', errors: [{ messageId: 'unexpected' }] },
    { code: 'for (eval in obj) {}', errors: [{ messageId: 'unexpected' }] },
    { code: 'export { eval }', errors: [{ messageId: 'unexpected' }] },
    { code: 'export { eval as foo }', errors: [{ messageId: 'unexpected' }] },
    { code: 'export default eval', errors: [{ messageId: 'unexpected' }] },
    { code: "eval['foo']", errors: [{ messageId: 'unexpected' }] },
    { code: 'eval.foo', errors: [{ messageId: 'unexpected' }] },
    { code: 'var obj = { key: eval }', errors: [{ messageId: 'unexpected' }] },
    { code: 'new eval()', errors: [{ messageId: 'unexpected' }] },

    // ================================================================
    // this.eval — top level of script (this IS global)
    // ================================================================
    { code: "this.eval('foo')", errors: [{ messageId: 'unexpected' }] },
    {
      code: "'use strict'; this.eval('foo')",
      errors: [{ messageId: 'unexpected' }],
    },
    // Bracket notation
    { code: "this['eval']('foo')", errors: [{ messageId: 'unexpected' }] },
    { code: "this[`eval`]('foo')", errors: [{ messageId: 'unexpected' }] },

    // ================================================================
    // this.eval — non-strict function (this defaults to global)
    // ================================================================
    {
      code: "function foo() { this.eval('foo') }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "function foo() { ('use strict'); this.eval; }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "async function foo() { this.eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "function* gen() { this.eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // isDefaultThisBinding — this IS default
    // ================================================================
    // Lowercase variable → not constructor
    {
      code: "var foo = function() { this.eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },
    // Named function → constructor heuristic doesn't apply
    {
      code: "var Foo = function foo() { this.eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },
    // .call/.apply with null/undefined/no args
    {
      code: "(function() { this.eval('foo') }).call(null)",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "(function() { this.eval('foo') }).call(undefined)",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "(function() { this.eval('foo') }).call()",
      errors: [{ messageId: 'unexpected' }],
    },
    // Returned function from non-IIFE
    {
      code: "function foo() { return function() { this.eval('bar') } }",
      errors: [{ messageId: 'unexpected' }],
    },
    // Callback without thisArg
    {
      code: "arr.forEach(function() { this.eval('foo') })",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'arr.flatMap(function(x) { return this.eval(x) })',
      errors: [{ messageId: 'unexpected' }],
    },
    // Callback with null thisArg
    {
      code: "arr.forEach(function() { this.eval('foo') }, null)",
      errors: [{ messageId: 'unexpected' }],
    },
    // reduce not a thisArg method
    {
      code: "arr.reduce(function(a, b) { return this.eval(a) ? a : b }, '0')",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Arrow functions — this inherits from enclosing scope
    // ================================================================
    {
      code: "() => { this.eval('foo'); }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "() => { 'use strict'; this.eval('foo'); }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "'use strict'; () => { this.eval('foo'); }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "() => { 'use strict'; () => { this.eval('foo'); } }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "function foo() { () => this.eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "function foo() { () => () => this.eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "var obj = { foo: () => this.eval('bar') }",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // this.eval in reference context
    // ================================================================
    {
      code: "var EVAL = this.eval; EVAL('foo')",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "'use strict'; var EVAL = this.eval; EVAL('foo')",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Optional chaining (using globalThis — window not in JS test env)
    // ================================================================
    { code: "globalThis?.eval('foo')", errors: [{ messageId: 'unexpected' }] },
    { code: "this?.eval('foo')", errors: [{ messageId: 'unexpected' }] },

    // ================================================================
    // Computed property name in class — this is outer scope
    // ================================================================
    {
      code: "class C { [this.eval('foo')] = 0 }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "'use strict'; class C { [this.eval('foo')] = 0 }",
      errors: [{ messageId: 'unexpected' }],
    },

    // ================================================================
    // Multiple violations
    // ================================================================
    {
      code: "eval('foo'); globalThis.eval('bar')",
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: "eval('a'); eval('b'); eval('c')",
      errors: [
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
      ],
    },
  ],
});
