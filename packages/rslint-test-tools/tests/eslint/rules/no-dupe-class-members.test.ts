import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-dupe-class-members', {
  valid: [
    'class A { foo() {} bar() {} }',
    'class A { static foo() {} foo() {} }',
    'class A { get foo() {} set foo(value) {} }',
    'class A { static foo() {} get foo() {} set foo(value) {} }',
    'class A { foo() { } } class B { foo() { } }',
    'class A { [foo]() {} foo() {} }',
    "class A { 'foo'() {} 'bar'() {} baz() {} }",
    "class A { *'foo'() {} *'bar'() {} *baz() {} }",
    "class A { get 'foo'() {} get 'bar'() {} get baz() {} }",
    'class A { 1() {} 2() {} }',
    "class A { ['foo']() {} ['bar']() {} }",
    'class A { [`foo`]() {} [`bar`]() {} }',
    'class A { [12]() {} [123]() {} }',
    "class A { [1.0]() {} ['1.0']() {} }",
    'class A { [0x1]() {} [`0x1`]() {} }',
    "class A { [null]() {} ['']() {} }",
    "class A { get ['foo']() {} set ['foo'](value) {} }",
    "class A { ['foo']() {} static ['foo']() {} }",
    // computed "constructor" key doesn't create a constructor
    "class A { ['constructor']() {} constructor() {} }",
    "class A { 'constructor'() {} [`constructor`]() {} }",
    'class A { constructor() {} get [`constructor`]() {} }',
    "class A { 'constructor'() {} set ['constructor'](value) {} }",
    // not assumed to be statically-known values
    "class A { ['foo' + '']() {} ['foo']() {} }",
    "class A { [`foo${''}`]() {} [`foo`]() {} }",
    "class A { [-1]() {} ['-1']() {} }",
    // non-literal computed key (not statically analyzed)
    'class A { [foo]() {} [foo]() {} }',
    // private and public
    'class A { foo; static foo; }',
    'class A { foo; #foo; }',
    "class A { '#foo'; #foo; }",
    // TypeScript method overloads
    'class Foo { foo(a: string): string; foo(a: number): number; foo(a: any): any {} }',
    'class A { static foo(a: string): void; static foo(a: number): void; static foo(a: any) {} }',
    'class A { constructor(a: string); constructor(a: number); constructor(a: any) {} }',
    'class A { foo(a: string): void; get foo() { return ""; } }',
    // abstract methods (no body)
    'abstract class A { abstract foo(): void; foo() {} }',
    'abstract class A { abstract get foo(): string; abstract set foo(v: string); }',
    // nested classes: outer and inner don't conflict
    'class Outer { foo() {} bar() { class Inner { foo() {} } } }',
    'class L1 { foo() {} bar() { class L2 { foo() {} baz() { class L3 { foo() {} } } } } }',
    // extends: same name across parent and child is fine
    'class Base { foo() {} } class Child extends Base { foo() {} }',
    // property with function-expression initializer
    'class A { foo = () => {} }',
    // mixed async / generator / accessor
    'class A { async foo() {} bar() {} }',
    'class A { *foo() {} bar() {} }',
    'class A { async *foo() {} bar() {} }',
    // index signature is not a class member for this rule
    'class A { [key: string]: any; foo() {} bar() {} }',
    // declare property alone
    'class A { declare foo: string; }',
    // empty / single-member class
    'class A {}',
    // non-static string-literal 'constructor' is still a constructor → skipped
    "class A { 'constructor'() {} constructor() {} }",
    // static + non-static constructor don't conflict (static is a method)
    'class A { static constructor() {} constructor() {} }',
    // Symbol-based computed key: dynamic, not deduplicated
    'class A { [Symbol.iterator]() {} [Symbol.hasInstance]() {} }',
    // side-effect expressions: not statically analyzed
    'class A { [a++]() {} [a++]() {} }',
  ],
  invalid: [
    {
      code: 'class A { foo() {} foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '!class A { foo() {} foo() {} };',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { 'foo'() {} 'foo'() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { 10() {} 1e1() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { ['foo']() {} ['foo']() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { static ['foo']() {} static foo() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { set 'foo'(value) {} set ['foo'](val) {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { ''() {} ['']() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { [`foo`]() {} [`foo`]() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { static get [`foo`]() {} static get ['foo']() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { foo() {} [`foo`]() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { get [`foo`]() {} 'foo'() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { static 'foo'() {} static [`foo`]() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    // computed 'constructor' duplicates (not the keyword constructor)
    {
      code: "class A { ['constructor']() {} ['constructor']() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { static [`constructor`]() {} static constructor() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { static constructor() {} static 'constructor'() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    // numeric literal equivalence
    {
      code: 'class A { [123]() {} [123]() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { [0x10]() {} 16() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { [100]() {} [1e2]() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { [123.00]() {} [`123`]() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { static '65'() {} static [0o101]() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { [123n]() {} 123() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "class A { [null]() {} 'null'() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    // triple duplicate
    {
      code: 'class A { foo() {} foo() {} foo() {} }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: 'class A { static foo() {} static foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // method / accessor cross-kind conflicts
    {
      code: 'class A { foo() {} get foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { set foo(value) {} foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // property declarations
    {
      code: 'class A { foo; foo; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // TS-specific: property + method
    {
      code: 'class A { foo; foo = 42; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { foo; foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // Overload + two implementations → duplicate only on the extra impl
    {
      code: 'class A { foo(a: string): void; foo(a: any) {} foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // async / generator duplicates
    {
      code: 'class A { async foo() {} async foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { *foo() {} *foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { foo() {} async foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // class expression with duplicates
    {
      code: 'var x = class { foo() {} foo() {} };',
      errors: [{ messageId: 'unexpected' }],
    },
    // nested classes: duplicate inside the inner class only
    {
      code: 'class Outer { bar() {} foo() { class Inner { baz() {} baz() {} } } }',
      errors: [{ messageId: 'unexpected' }],
    },
    // duplicates in both outer and inner class
    {
      code: 'class Outer { foo() {} foo() {} bar() { class Inner { baz() {} baz() {} } } }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    // state isolation: independent names reported independently
    {
      code: 'class A { foo() {} bar() {} foo() {} bar() {} }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    // init taints both accessors → 2 errors
    {
      code: 'class A { foo = 1; get foo() {} set foo(v) {} }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    // duplicate getter followed by valid setter → 1 error (getter), setter still pairs
    {
      code: 'class A { get foo() {} get foo() {} set foo(v) {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // method + dup getter + setter → 3 errors
    {
      code: 'class A { foo() {} get foo() {} get foo() {} set foo(v) {} }',
      errors: [
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
      ],
    },
    // extends: child duplicates still detected
    {
      code: 'class Base { foo() {} } class Child extends Base { foo() {} foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // __proto__ as class method: duplicates detected
    {
      code: 'class A { __proto__() {} __proto__() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // property with arrow function initializer
    {
      code: 'class A { foo = () => {}; foo = () => {}; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // index signature present, duplicates elsewhere still caught
    {
      code: 'class A { [key: string]: any; foo() {} foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // definite assignment (!) doesn't affect the name
    {
      code: 'class A { foo!: string; foo!: number; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // access modifiers don't affect name
    {
      code: 'class A { private foo() {} public foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // readonly doesn't affect name
    {
      code: 'class A { readonly foo: string; foo: number; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // optional property doesn't affect name
    {
      code: 'class A { foo?: string; foo: number; }',
      errors: [{ messageId: 'unexpected' }],
    },
    // PropertyDeclaration with computed name
    {
      code: "class A { ['foo'] = 1; foo() {} }",
      errors: [{ messageId: 'unexpected' }],
    },
    // declare property + method
    {
      code: 'class A { declare foo: string; foo() {} }',
      errors: [{ messageId: 'unexpected' }],
    },
    // static and non-static each independently duplicated
    {
      code: 'class A { foo() {} static foo() {} foo() {} static foo() {} }',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
  ],
});
