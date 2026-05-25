import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-new', {
  valid: [
    // ---- Original ESLint cases ----
    'var a = new Date()',
    'var a; if (a === new Date()) { a = false; }',
    // ---- Used as a value ----
    'const thing = new Thing();',
    'let x = new Foo();',
    'foo(new Bar());',
    'function f() { return new Baz(); }',
    'var x = new A() || new B();',
    'var x = true ? new A() : new B();',
    'var x = [new A(), new B()];',
    'var x = { a: new A() };',
    'class C { m() { return new A(); } }',
    '(function () { return new Foo(); })();',
    // ---- Result further dereferenced / called ----
    '(new Foo).bar;',
    '(new Foo()).bar();',
    'new Foo().bar;',
    'new Foo()?.bar;',
    'new Foo()[0];',
    'new Foo().bar();',
    // ---- Wrapped by another operator ----
    'void new Foo();',
    '!new Foo();',
    'typeof new Foo();',
    'delete new Foo().x;',
    'new Foo(), new Bar();',
    'async function f() { await new Promise(r => r(1)); }',
    'function* g() { yield new Foo(); }',
    // ---- Plain call ----
    'Foo();',
    'foo.bar();',
    // ---- Arrow expression body ----
    'var f = () => new Foo();',
    // ---- Wrapping operators make the statement expression non-NewExpression ----
    'true && new Foo();',
    'a ? new Foo() : new Bar();',
    // ---- TypeScript type casts / satisfies around the new ----
    'new Foo() as Bar;',
    '(new Foo() as Bar);',
    '<Foo>new Bar();',
    'new Foo() satisfies Bar;',
    // ---- Not an ExpressionStatement ----
    'export default new Foo();',
    '@new Dec() class C {}',
    // ---- Class field / default parameter initializers ----
    'class C { x = new Foo(); }',
    'class C { static x = new Foo(); }',
    'function f(x = new Foo()) {}',
    // ---- Chained type casts ----
    'new Foo() as any as Bar;',
    // ---- Destructuring default values / computed keys ----
    'var { a = new Foo() } = {};',
    'var [a = new Foo()] = [];',
    'var obj = { [new Foo()]: 1 };',
    // ---- Template literal expression slot ----
    '`${new Foo()}`;',
    // ---- Calling the result of new ----
    '(new Foo())();',
    // ---- TC39 accessor keyword ----
    'class C { accessor x = new Foo(); }',
    // ---- Decorator factory call ----
    '@factory(new Foo()) class C {}',
    // ---- import.meta call ----
    'import.meta.foo();',
    // ---- `using` / `await using` declarations ----
    'function f() { using x = new Foo(); }',
    'async function f() { await using x = new Foo(); }',
    // ---- Chained / combined TS type operators ----
    'new Foo() as const;',
    '<const>new Foo();',
    '(new Foo())!;',
    // ---- Non-null or type-assertion on callee, used as value ----
    'var x = new Foo!();',
    'var x = new (Foo as any)();',
    // ---- Decorators as values / on methods ----
    '@a @b() @c(new X()) class C {}',
    'class C { @dec method() { return new Foo(); } }',
    // ---- implements / mixin extends (as value) ----
    'class C implements I { m() { return new Foo(); } }',
    'class C extends A implements I { m() { return new Foo(); } }',
    // ---- Generator yield / yield* new ----
    'function* g() { yield* new Foo(); }',
    'async function* g() { yield new Foo(); }',
    // ---- Multi-generic / member-access generic ----
    'var x = new Foo<number, string>();',
    'var x = new ns.Foo<number>();',
  ],
  invalid: [
    // ---- Original ESLint case ----
    {
      code: 'new Date()',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Basic forms ----
    {
      code: 'new Foo;',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: "new Foo('a', 1);",
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'new foo.Bar();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'new foo.bar.Baz();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Parenthesized: ESLint flags (paren-transparent) ----
    {
      code: '(new Foo());',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: '((new Foo()));',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- new of new ----
    {
      code: 'new new Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Labeled statement ----
    {
      code: 'label: new Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Nested inside various constructs ----
    {
      code: 'function f() { new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'var f = () => { new Foo(); };',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'if (true) { new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'class C { m() { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'for (;;) { new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'try { new Foo(); } catch {}',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Multiple / multi-line ----
    {
      code: 'new Foo(); new Bar();',
      errors: [
        { messageId: 'noNewStatement' },
        { messageId: 'noNewStatement' },
      ],
    },
    {
      code: 'new\n  Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- TypeScript generics on callee ----
    {
      code: 'new Foo<number>();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Class-body static block ----
    {
      code: 'class C { static { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Anonymous class as constructor ----
    {
      code: 'new class {}();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Tagged-template as new callee ----
    {
      code: 'new Foo``;',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- More container scopes ----
    {
      code: 'do { new Foo(); } while (false);',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'while (true) { new Foo(); break; }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'switch (x) { case 1: new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'if (a) {} else { new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'try {} finally { new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: '{ new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Comments around / inside the expression ----
    {
      code: '/* a */ new Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'new /* a */ Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Single-statement branches (no block) ----
    {
      code: 'if (a) new Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'if (a) new Foo(); else new Bar();',
      errors: [
        { messageId: 'noNewStatement' },
        { messageId: 'noNewStatement' },
      ],
    },
    {
      code: 'for (var i = 0; i < 1; i++) new Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'while (a) new Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- IIFE-style: callee is a parenthesized function expression ----
    {
      code: 'new (function(){})();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Spread argument ----
    {
      code: 'new Foo(...args);',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Computed-access callee (no call parens) ----
    {
      code: 'new Foo[0];',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- ASI splits into two statements ----
    {
      code: 'var a = b\nnew Foo()',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Trailing extra semicolon is an EmptyStatement ----
    {
      code: 'new Foo();;',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- TypeScript namespace body ----
    {
      code: 'namespace N { new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Class expression as constructor ----
    {
      code: 'new (class {})();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'new (class extends Base {})();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Async IIFE ----
    {
      code: 'new (async function(){})();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Computed-member callee ----
    {
      code: 'new obj[key]();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Deep member chain ----
    {
      code: 'new a.b.c.d();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Tagged template with member-access tag ----
    {
      code: 'new foo.Bar`x`;',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Double report ----
    {
      code: 'new Promise(r => { new Foo(); r(); });',
      errors: [
        { messageId: 'noNewStatement' },
        { messageId: 'noNewStatement' },
      ],
    },
    // ---- After directive prologue ----
    {
      code: '"use strict"; new Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Interleaved between other statements ----
    {
      code: 'foo; new Bar(); baz;',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Class body: constructor / derived / static / getter / setter / private method ----
    {
      code: 'class C { constructor() { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'class C extends B { constructor() { super(); new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'class C { static m() { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'class C { get x() { new Foo(); return 1; } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'class C { set x(v) { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'class C { #m() { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Object literal method / accessor bodies ----
    {
      code: 'var obj = { m() { new Foo(); } };',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'var obj = { get x() { new Foo(); return 1; } };',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'var obj = { set x(v) { new Foo(); } };',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Catch / for-in / for-of / for-await-of ----
    {
      code: 'try {} catch (e) { new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'for (var x in arr) new Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'for (var x of arr) new Foo();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'async function f() { for await (const x of arr) new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Async IIFE with inner new statement ----
    {
      code: '(async () => { new Foo(); })();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Outer new-statement AND inner new inside an argument ----
    {
      code: 'new Foo(function() { new Bar(); });',
      errors: [
        { messageId: 'noNewStatement' },
        { messageId: 'noNewStatement' },
      ],
    },
    // ---- Nested namespace ----
    {
      code: 'namespace A.B { new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Legacy `module` keyword ----
    {
      code: 'module N { new Foo(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Abstract class constructor ----
    {
      code: 'abstract class A { constructor() { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- `override` method body ----
    {
      code: 'class B extends A { override m() { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Decorated method body ----
    {
      code: 'class C { @dec m() { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Class with implements clause ----
    {
      code: 'class C implements I { m() { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Mixin-style extends ----
    {
      code: 'class C extends Mix(Base) { m() { new Foo(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Static block with mixed decl + new statement ----
    {
      code: 'class C { static { let x = new Foo(); new Bar(); } }',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Multi-generic / member-access generic ----
    {
      code: 'new Foo<number, string>();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    {
      code: 'new ns.Foo<number>();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Non-null on callee ----
    {
      code: 'new Foo!();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Type assertion on callee ----
    {
      code: 'new (Foo as any)();',
      errors: [{ messageId: 'noNewStatement' }],
    },
    // ---- Dynamic import awaited then new'd ----
    {
      code: 'async function f() { new (await import("x")).default(); }',
      errors: [{ messageId: 'noNewStatement' }],
    },
  ],
});
