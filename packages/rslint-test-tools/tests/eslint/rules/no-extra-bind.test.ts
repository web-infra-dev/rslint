import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-extra-bind', {
  valid: [
    // --- Bind access patterns ---
    'var a = function(b: any) { return b }.bind(c, d)',
    'var a = function(b: any) { return b }.bind(...c)',
    'var a = function() { return 1; }.bind()',
    'var a = function() { return 1; }[bind](b)',
    'var a = (() => { return b }).bind(c, d)',

    // --- Not .bind() ---
    'var a = f.bind(a)',
    'f.bind(a)',
    '(function() { this.b; }).call(c)',
    '(function() { return 1; }).apply(c)',
    'var a = function() { this.b }()',
    'var a = function() { this.b }.foo()',
    'var a = function() { return 1; }',

    // --- Function uses this directly ---
    'var a = function() { this.b }.bind(c)',
    'var a = function() { return this; }.bind(c)',
    'var a = function() { this.b; return 1; }.bind(c)',

    // --- Arrow captures this from outer scope ---
    'var a = function() { return () => this; }.bind(b)',
    'var a = function() { var f = () => this }.bind(c)',
    'var a = function() { var f = () => () => this }.bind(c)',

    // --- Nested bind where outer uses this ---
    '(function() { (function() { this.b }.bind(this)) }.bind(c))',

    // --- This in function + also in class method (direct this counts) ---
    'var a = function() { this.x; class Foo { bar() { this.y } } }.bind(c)',

    // --- Computed property name: this in [this.key]() belongs to outer scope ---
    'var a = function() { var o = { [this.key]() {} } }.bind(c)',
    'var a = function() { class Foo { [this.key]() {} } }.bind(c)',
    'var a = function() { var o = { [this.key]: 1 } }.bind(c)',
    'var a = function() { class Foo { x = this.y } }.bind(c)',
    'var a = function(a = this) { return a }.bind(c)',
    // Computed getter/setter names
    'var a = function() { var o = { get [this.key]() { return 1 } } }.bind(c)',
    'var a = function() { var o = { set [this.key](v) {} } }.bind(c)',

    // --- this in extends clause belongs to outer scope ---
    'var a = function() { class Foo extends this.Base {} }.bind(c)',

    // --- this in arrow default parameter inherits outer this ---
    'var a = function() { var f = (x = this) => x }.bind(c)',

    // --- Triple nested arrow chain inherits this ---
    'var a = function() { var f = () => () => () => this }.bind(c)',

    // --- this in computed property name inside arrow (arrow transparent) ---
    'var a = function() { var f = () => { class Foo { [this.key]() {} } } }.bind(c)',

    // --- this in class static initializer block (not a this-scope boundary) ---
    'var a = function() { class Foo { static { this.x = 1 } } }.bind(c)',

    // --- this in try/catch/finally (no scope boundary) ---
    'var a = function() { try {} catch(e) { this.x } }.bind(c)',
    'var a = function() { try {} finally { this.x } }.bind(c)',

    // --- this in extends + static field together ---
    'var a = function() { class Foo extends this.Base { static x = this } }.bind(c)',
  ],
  invalid: [
    // --- Basic: function doesn't use this ---
    {
      code: 'var a = function() { return 1; }.bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { return 1; }.bind(this)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Arrow function: .bind() is always unnecessary ---
    {
      code: 'var a = (() => { return 1; }).bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = (() => { this.b }).bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = (() => { return this; }).bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = (() => 1).bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- this in nested function does NOT count for outer ---
    {
      code: 'var a = function() { (function(){ this.c }) }.bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { function inner() { this.b; } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { function c(){ this.d } }.bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Bind access patterns ---
    {
      code: "var a = function() { return 1; }['bind'](b)",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { return 1; }[`bind`](b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = (function() { return 1; }.bind)(this)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = (function() { return 1; }).bind(this)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Parenthesized: multiple levels ---
    {
      code: 'var a = ((function() { return 1; })).bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = ((function() { return 1; }.bind))(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = ((function() { return 1; }).bind)(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "var a = (function() { return 1; }['bind'])(c)",
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Deeply nested ---
    {
      code: 'var a = function() { (function(){ (function(){ this.d }.bind(c)) }) }.bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { function a() { function b() { this.x } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- this in function inside arrow inside bound function ---
    {
      code: 'var a = function() { var f = () => { return function() { this.x } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    // --- this in arrow inside inner function (arrow this = inner fn's this) ---
    {
      code: 'var a = function() { function inner() { var f = () => this } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Async / Generator ---
    {
      code: 'var a = (async function() { return 1; }).bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = (function*() { yield 1; }).bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = (async () => { return 1; }).bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- this-scoping: class methods/accessors/constructors isolate this ---
    {
      code: 'var a = function() { class Foo { bar() { this.x } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { var o = { foo() { this.x } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { var o = { get foo() { return this.x } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { var o = { set foo(v) { this.x = v } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { class Foo { constructor() { this.x = 1 } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() { class Foo { bar() { var f = () => this } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- this in default param of inner function belongs to inner scope ---
    {
      code: 'var a = function() { function inner(a = this) {} }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    // --- this in property value (FunctionExpression isolates) ---
    {
      code: 'var a = function() { var o = { foo: function() { this.x } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    // --- this in computed name of nested method belongs to containing method ---
    {
      code: 'var a = function() { var o = { foo() { var p = { [this.key]() {} } } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Static method isolates this ---
    {
      code: 'var a = function() { class Foo { static bar() { this.x } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    // --- Class expression: method isolates this ---
    {
      code: 'var a = function() { var C = class { method() { this.x } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    // --- Method inside arrow (arrow transparent, method isolates) ---
    {
      code: 'var a = function() { var f = () => ({ method() { this.x } }) }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    // --- Nested class constructors both isolate ---
    {
      code: 'var a = function() { class Foo { constructor() { class Bar { constructor() { this.x } } } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    // --- Deeply nested arrows inside method (all arrows transparent, method isolates) ---
    {
      code: 'var a = function() { class Foo { bar() { var f = () => { var g = () => this } } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },
    // --- this in computed name of nested class method (belongs to outer method scope) ---
    {
      code: 'var a = function() { class Outer { method() { class Inner { [this.key]() {} } } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Comment preservation in autofix ---
    {
      code: 'var a = function() {}/**/.bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() {} // comment\n.bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() {} /* a */ /* b */ .bind(b)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Named function expression ---
    {
      code: 'var a = (function foo() { return 1; }).bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Deep alternating nesting: function -> method -> arrow -> function -> constructor ---
    {
      code: 'var a = function() { class Foo { method() { var f = () => { var g = function() { class Bar { constructor() { this.x } } } } } } }.bind(c)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Multiple bind calls: both reported ---
    {
      code: 'var a = function() { var b = function() { return 1 }.bind(d) }.bind(c)',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },

    // --- Autofix with null arg ---
    {
      code: 'var a = function() {}.bind(null)',
      errors: [{ messageId: 'unexpected' }],
    },

    // --- Side-effect arguments: no autofix ---
    {
      code: 'var a = function() {}.bind(b++)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() {}.bind(b())',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() {}.bind(b.c)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() {}.bind(`${b}`)',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() {}.bind([])',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'var a = function() {}.bind({})',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});
