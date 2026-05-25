import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Mirrors the Go test file 1:1. Framework-level ESLint concepts
// (languageOptions.globals, parserOptions.ecmaFeatures.globalReturn, env,
// sourceType: "script" vs "module" distinction) are not exercised here; see
// the `SKIP` comments in the Go file for the full list.
ruleTester.run('no-shadow', {
  valid: [
    // ---- Baseline (no shadow) ----
    'var a=3; function b(x) { a++; return x + a; }; setTimeout(function() { b(a); }, 0);',

    // ---- Function-name initializer exception (function expression) ----
    '(function() { var doSomething = function doSomething() {}; doSomething() }())',
    '(function() { var doSomething = foo || function doSomething() {}; doSomething() }())',
    '(function() { var doSomething = function doSomething() {} || foo; doSomething() }())',
    '(function() { var doSomething = foo && function doSomething() {}; doSomething() }())',
    '(function() { var doSomething = foo ?? function doSomething() {}; doSomething() }())',
    '(function() { var doSomething = foo || (bar || function doSomething() {}); doSomething() }())',
    '(function() { var doSomething = foo || (bar && function doSomething() {}); doSomething() }())',
    '(function() { var doSomething = foo ? function doSomething() {} : bar; doSomething() }())',
    '(function() { var doSomething = foo ? bar: function doSomething() {}; doSomething() }())',
    '(function() { var doSomething = foo ? bar: (baz || function doSomething() {}); doSomething() }())',
    '(function() { var doSomething = (foo ? bar: function doSomething() {}) || baz; doSomething() }())',
    '(function() { var { doSomething = function doSomething() {} } = obj; doSomething() }())',
    '(function() { var { doSomething = function doSomething() {} || foo } = obj; doSomething() }())',
    '(function() { var { doSomething = foo ? function doSomething() {} : bar } = obj; doSomething() }())',
    '(function() { var { doSomething = foo ? bar : function doSomething() {} } = obj; doSomething() }())',
    '(function() { var { doSomething = foo || (bar ? baz : (qux || function doSomething() {})) || quux } = obj; doSomething() }())',
    'function foo(doSomething = function doSomething() {}) { doSomething(); }',
    'function foo(doSomething = function doSomething() {} || foo) { doSomething(); }',
    'function foo(doSomething = foo ? function doSomething() {} : bar) { doSomething(); }',
    'function foo(doSomething = foo ? bar : function doSomething() {}) { doSomething(); }',
    'function foo(doSomething = foo || (bar ? baz : (qux || function doSomething() {})) || quux) { doSomething(); }',

    // ---- Miscellaneous ----
    'var arguments;\nfunction bar() { }',
    'var a=3; var b = (x) => { a++; return x + a; }; setTimeout(() => { b(a); }, 0);',

    // ---- Class (no shadow) and class-expression initializer exceptions ----
    'class A {}',
    'class A { constructor() { var a; } }',
    '(function() { var A = class A {}; })()',
    '(function() { var A = foo || class A {}; })()',
    '(function() { var A = class A {} || foo; })()',
    '(function() { var A = foo && class A {} || foo; })()',
    '(function() { var A = foo ?? class A {}; })()',
    '(function() { var A = foo || (bar || class A {}); })()',
    '(function() { var A = foo || (bar && class A {}); })()',
    '(function() { var A = foo ? class A {} : bar; })()',
    '(function() { var A = foo ? bar : class A {}; })()',
    '(function() { var A = foo ? bar: (baz || class A {}); })()',
    '(function() { var A = (foo ? bar: class A {}) || baz; })()',
    '(function() { var { A = class A {} } = obj; }())',
    '(function() { var { A = class A {} || foo } = obj; }())',
    '(function() { var { A = foo ? class A {} : bar } = obj; }())',
    '(function() { var { A = foo ? bar : class A {} } = obj; }())',
    '(function() { var { A = foo || (bar ? baz : (qux || class A {})) || quux } = obj; }())',
    'function foo(A = class A {}) { doSomething(); }',
    'function foo(A = class A {} || foo) { doSomething(); }',
    'function foo(A = foo ? class A {} : bar) { doSomething(); }',
    'function foo(A = foo ? bar : class A {}) { doSomething(); }',
    'function foo(A = foo || (bar ? baz : (qux || class A {})) || quux) { doSomething(); }',

    // ---- Block redecl (not shadow) ----
    '{ var a; } var a;',

    // ---- hoist default ("functions") ----
    '{ let a; } let a;',
    '{ let a; } var a;',
    '{ const a = 0; } const a = 1;',
    '{ const a = 0; } var a;',

    // ---- hoist: never ----
    { code: '{ let a; } let a;', options: [{ hoist: 'never' }] as any },
    { code: '{ let a; } var a;', options: [{ hoist: 'never' }] as any },
    {
      code: '{ let a; } function a() {}',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: '{ const a = 0; } const a = 1;',
      options: [{ hoist: 'never' }] as any,
    },
    { code: '{ const a = 0; } var a;', options: [{ hoist: 'never' }] as any },
    {
      code: '{ const a = 0; } function a() {}',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'function foo() { let a; } let a;',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'function foo() { let a; } var a;',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'function foo() { let a; } function a() {}',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'function foo() { var a; } let a;',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'function foo() { var a; } var a;',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'function foo() { var a; } function a() {}',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'function foo(a) { } let a;',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'function foo(a) { } var a;',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'function foo(a) { } function a() {}',
      options: [{ hoist: 'never' }] as any,
    },

    // ---- builtinGlobals default: off ----
    'function foo() { var Object = 0; }',

    // ---- allow list ----
    {
      code: 'function foo(cb) { (function (cb) { cb(42); })(cb); }',
      options: [{ allow: ['cb'] }] as any,
    },

    // ---- Class (property vs method same name) ----
    'class C { foo; foo() { let foo; } }',

    // ---- Class static blocks ----
    'class C { static { var x; } static { var x; } }',
    'class C { static { let x; } static { let x; } }',
    'class C { static { var x; { var x; /* redeclaration */ } } }',
    'class C { static { { var x; } { var x; /* redeclaration */ } } }',
    'class C { static { { let x; } { let x; } } }',

    // ---- ignoreOnInitialization (callback) ----
    {
      code: 'const a = [].find(a => a)',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'const a = [].find(function(a) { return a; })',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'const [a = [].find(a => true)] = dummy',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'const { a = [].find(a => true) } = dummy',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'function func(a = [].find(a => true)) {}',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'for (const a in [].find(a => true)) {}',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'for (const a of [].find(a => true)) {}',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: "const a = [].map(a => true).filter(a => a === 'b')",
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: "const a = [].map(a => true).filter(a => a === 'b').find(a => a === 'c')",
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'const { a } = (({ a }) => ({ a }))();',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: "const person = people.find(item => {const person = item.name; return person === 'foo'})",
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'var y = bar || foo(y => y);',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'var y = bar && foo(y => y);',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'var z = bar(foo(z => z));',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'var z = boo(bar(foo(z => z)));',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: "var match = function (person) { return person.name === 'foo'; };\nconst person = [].find(match);",
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'const a = foo(x || (a => {}))',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'const { a = 1 } = foo(a => {})',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: "const person = {...people.find((person) => person.firstName.startsWith('s'))}",
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: "const person = { firstName: people.filter((person) => person.firstName.startsWith('s')).map((person) => person.firstName)[0]}",
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: '() => { const y = foo(y => y); }',
      options: [{ ignoreOnInitialization: true }] as any,
    },

    // ---- ignoreOnInitialization (IIFE) ----
    {
      code: 'const x = (x => x)()',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'var y = bar || (y => y)();',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'var y = bar && (y => y)();',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'var x = (x => x)((y => y)());',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: 'const { a = 1 } = (a => {})()',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: '() => { const y = (y => y)(); }',
      options: [{ ignoreOnInitialization: true }] as any,
    },
    'const [x = y => y] = [].map(y => y)',

    // ---- TS function-type parameters (default ignoreFunctionTypeParameterNameValueShadow: true) ----
    'function foo<T = (arg: any) => any>(arg: T) {}',
    'function foo<T = ([arg]: [any]) => any>(arg: T) {}',
    'function foo<T = ({ args }: { args: any }) => any>(arg: T) {}',
    'function foo<T = (...args: any[]) => any>(fn: T, args: any[]) {}',
    'function foo<T extends (...args: any[]) => any>(fn: T, args: any[]) {}',
    'function foo<T extends (...args: any[]) => any>(fn: T, ...args: any[]) {}',
    'function foo<T extends ([args]: any[]) => any>(fn: T, args: any[]) {}',
    'function foo<T extends ([...args]: any[]) => any>(fn: T, args: any[]) {}',
    'function foo<T extends ({ args }: { args: any }) => any>(fn: T, args: any) {}',
    'function foo<T extends (id: string, ...args: any[]) => any>(fn: T, ...args: any[]) {}',
    'type Args = 1; function foo<T extends (Args: any) => void>(arg: T) {}',

    // ---- Conditional types with infer T ----
    'export type ArrayInput<Func> = Func extends (arg0: Array<infer T>) => any ? T[] : Func extends (...args: infer T) => any ? T : never;',

    // ---- Declaration merging ----
    'class Foo { prop = 1; } namespace Foo { export const v = 2; }',
    'function Foo() {} namespace Foo { export const v = 2; }',
    'class Foo { prop = 1; } interface Foo { prop2: string }',

    // ---- this param ----
    'function test(this: number) { function test2(this: number) {} }',

    // ---- Type-only import + module augmentation of same module ----
    "import type { Foo } from 'bar';\ndeclare module 'bar' { export interface Foo { x: string } }",

    // ---- TS type/value shadow (default ignoreTypeValueShadow: true) ----
    'const x = 1; type x = string;',
    'const x = 1; { type x = string; }',

    // ---- TS enum constant initializers ----
    "enum Direction { left = 'left', right = 'right' }",

    // ---- ignoreFunctionTypeParameterNameValueShadow: true (default) — various TS signature positions ----
    'const test = 1; type Fn = (test: string) => typeof test;',
    'const arg = 0; interface Test { (arg: string): typeof arg; }',
    'const arg = 0; interface Test { p1(arg: string): typeof arg; }',
    'const arg = 0; declare function test(arg: string): typeof arg;',
    'const arg = 0; declare const test: (arg: string) => typeof arg;',
    'const arg = 0; declare class Test { p1(arg: string): typeof arg; }',
    'const arg = 0; declare const Test: { new (arg: string): typeof arg };',
    'const arg = 0; type Bar = new (arg: number) => typeof arg;',
    'const arg = 0; declare namespace Lib { function test(arg: string): typeof arg; }',

    // ---- declare global is transparent ----
    {
      code: 'declare global { interface ArrayConstructor {} } export {};',
      options: [{ builtinGlobals: true }] as any,
    },
    'declare global { const a: string; namespace Foo { const a: number; } } export {};',
    {
      code: "declare global { type A = 'foo'; namespace Foo { type A = 'bar'; } } export {};",
      options: [{ ignoreTypeValueShadow: false }] as any,
    },

    // ---- Static vs instance class generic ----
    'export class Wrapper<Wrapped> { private constructor(private readonly wrapped: Wrapped) {} unwrap(): Wrapped { return this.wrapped; } static create<Wrapped>(wrapped: Wrapped) { return new Wrapper<Wrapped>(wrapped); } }',
    'function makeA() { return class A<T> { constructor(public value: T) {} static make<T>(value: T) { return new A<T>(value); } }; }',

    // ---- Import type + type alias — value shadow allowed under default ----
    "import type { foo } from './foo';\ntype bar = number;\nfunction doThing(foo: number, bar: number) {}",
    "import { type foo } from './foo';\nfunction doThing(foo: number) {}",

    // ---- TS hoist: never ----
    {
      code: 'type Foo<A> = 1; type A = 1;',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'interface Foo<A> {} type A = 1;',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'interface Foo<A> {} interface A {}',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: 'type Foo<A> = 1; interface A {}',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: '{ type A = 1; } type A = 1;',
      options: [{ hoist: 'never' }] as any,
    },
    {
      code: '{ interface Foo<A> {} } type A = 1;',
      options: [{ hoist: 'never' }] as any,
    },

    // ---- TS hoist: functions (type not reported under this mode) ----
    {
      code: 'type Foo<A> = 1; type A = 1;',
      options: [{ hoist: 'functions' }] as any,
    },
    {
      code: 'interface Foo<A> {} type A = 1;',
      options: [{ hoist: 'functions' }] as any,
    },
    {
      code: 'interface Foo<A> {} interface A {}',
      options: [{ hoist: 'functions' }] as any,
    },
    {
      code: 'type Foo<A> = 1; interface A {}',
      options: [{ hoist: 'functions' }] as any,
    },
    {
      code: '{ type A = 1; } type A = 1;',
      options: [{ hoist: 'functions' }] as any,
    },
    {
      code: '{ interface Foo<A> {} } type A = 1;',
      options: [{ hoist: 'functions' }] as any,
    },

    // ---- Real-code-discovered: cross-specifier import-type quirk ----
    "import binding, { type AssetInfo } from 'm';\nclass Foo { static __from_binding(binding: any) { return binding; } }",
    "import { foo, type Bar } from 'm';\nfunction fn(foo: number) { return foo; }",

    // ---- infer X is type-level only ----
    'type X<T> = T extends infer U ? U : never;\nconst U = 1;',
    'type X<T> = T extends string ? T : never;\nconst T = 1;',

    // ---- Object literal accessor distinct names ----
    'const x = 1; const o = { get y() { return 1; }, set y(v) {} };',

    // ---- Class private field name ----
    'class C { #priv = 1; m() { const priv = 2; return [this.#priv, priv]; } }',

    // ---- Tuple labels not bindings ----
    'type Pair = [first: string, second: number];\nconst first = 1;',

    // ---- Class field name same as class ----
    'class C6 { C6 = 1; }',

    // ---- Method name same as outer ----
    'function g1() {}\nclass CG { g1() {} }',

    // ---- Mapped type binding doesn't leak ----
    'type Map1<T> = { [K in keyof T]: K };\nconst K = 1;',

    // ---- Generic in arrow ----
    'const T = 1; const arr = <T extends string>(x: T) => x;',

    // ---- Multiple sibling `infer U` at same level ----
    'type X<T> = T extends { a: infer U } & { b: infer U } ? U : never;',

    // ---- Module augmentation (valid forms) ----
    "import type { Foo } from 'bar';\ndeclare module 'bar' { export type Foo = string }",
    "import type { Foo } from 'bar';\ndeclare module 'bar' { interface Foo { x: string } }",
    "import { type Foo } from 'bar';\ndeclare module 'bar' { export type Foo = string }",
    "import { type Foo } from 'bar';\ndeclare module 'bar' { export interface Foo { x: string } }",
    "import { type Foo } from 'bar';\ndeclare module 'bar' { type Foo = string }",
    "import { type Foo } from 'bar';\ndeclare module 'bar' { interface Foo { x: string } }",

    // ==== Additional cases mirroring Go test additions ====
    {
      code: `var arguments;
function bar() { }`,
    },
    {
      code: `var match = function (person) { return person.name === 'foo'; };
const person = [].find(match);`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `import type { Foo } from 'bar';
declare module 'bar' { export interface Foo { x: string } }`,
    },
    {
      code: `import binding, { type AssetInfo } from 'm';
class Foo { static __from_binding(binding: any) { return binding; } }`,
    },
    {
      code: `import { foo, type Bar } from 'm';
function fn(foo: number) { return foo; }`,
    },
    {
      code: `type X<T> = T extends infer U ? U : never;
const U = 1;`,
    },
    {
      code: `type X<T> = T extends string ? T : never;
const T = 1;`,
    },
    {
      code: `type Pair = [first: string, second: number];
const first = 1;`,
    },
    {
      code: `function g1() {}
class CG { g1() {} }`,
    },
    {
      code: `type Map1<T> = { [K in keyof T]: K };
const K = 1;`,
    },
    {
      code: `import type { foo } from './foo';
type bar = number;
function doThing(foo: number, bar: number) {}`,
    },
    {
      code: `import { type foo } from './foo';
function doThing(foo: number) {}`,
    },
    {
      code: `const a = [].find(a => a);`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `const a = [].find(function (a) { return a; });`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `const [a = [].find(a => true)] = dummy;`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `const { a = [].find(a => true) } = dummy;`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `import type { Foo } from 'bar';
declare module 'bar' { export type Foo = string }`,
    },
    {
      code: `import type { Foo } from 'bar';
declare module 'bar' { interface Foo { x: string } }`,
    },
    {
      code: `import { type Foo } from 'bar';
declare module 'bar' { export type Foo = string }`,
    },
    {
      code: `import { type Foo } from 'bar';
declare module 'bar' { export interface Foo { x: string } }`,
    },
    {
      code: `import { type Foo } from 'bar';
declare module 'bar' { type Foo = string }`,
    },
    {
      code: `import { type Foo } from 'bar';
declare module 'bar' { interface Foo { x: string } }`,
    },
    {
      code: `var match = function (person) { return person.name === 'foo'; };
const person = [].find(match);`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `
  const arg = 0;
  
  declare const Test: {
	new (arg: string): typeof arg;
  };
		`,
      options: [{ ignoreFunctionTypeParameterNameValueShadow: true }] as any,
    },
    {
      code: `
		  declare global {
			const foo: string;
			type Fn = (foo: number) => void;
		  }
		  export {};
		`,
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
    },
    {
      code: `
  import { type foo } from './foo';
  
  // 'foo' is already declared in the upper scope
  function doThing(foo: number) {}
		`,
      options: [{ ignoreTypeValueShadow: true }] as any,
    },
    {
      code: `const a = [].map(a => true).filter(a => a === 'b');`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `
  const a = []
	.map(a => true)
	.filter(a => a === 'b')
	.find(a => a === 'c');
		`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `
  const person = people.find(item => {
	const person = item.name;
	return person === 'foo';
  });
		`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `
  var match = function (person) {
	return person.name === 'foo';
  };
  const person = [].find(match);
		`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `const a = foo(x || (a => {}));`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `const { a = 1 } = foo(a => {});`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `const person = { ...people.find(person => person.firstName.startsWith('s')) };`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `
  const person = {
	firstName: people
	  .filter(person => person.firstName.startsWith('s'))
	  .map(person => person.firstName)[0],
  };
		`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `
  () => {
	const y = foo(y => y);
  };
		`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `const x = (x => x)();`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `const { a = 1 } = (a => {})();`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `
  () => {
	const y = (y => y)();
  };
		`,
      options: [{ ignoreOnInitialization: true }] as any,
    },
    {
      code: `var arguments;
function bar() { }`,
    },
    {
      code: `
  function test(this: Foo) {
	function test2(this: Bar) {}
  }
	  `,
    },
    {
      code: `
  class Foo {
	prop = 1;
  }
  interface Foo {
	prop2: string;
  }
	  `,
    },
    {
      code: `
  import type { Foo } from 'bar';
  
  declare module 'bar' {
	export interface Foo {
	  x: string;
	}
  }
	  `,
    },
    {
      code: `
  enum Direction {
	left = 'left',
	right = 'right',
  }
	  `,
    },
    {
      code: `const [x = y => y] = [].map(y => y);`,
    },
    {
      code: `
  import type { Foo } from 'bar';
  
  declare module 'bar' {
	export type Foo = string;
  }
		`,
    },
    {
      code: `
  import type { Foo } from 'bar';
  
  declare module 'bar' {
	interface Foo {
	  x: string;
	}
  }
		`,
    },
    {
      code: `
  import { type Foo } from 'bar';
  
  declare module 'bar' {
	export type Foo = string;
  }
		`,
    },
    {
      code: `
  import { type Foo } from 'bar';
  
  declare module 'bar' {
	export interface Foo {
	  x: string;
	}
  }
		`,
    },
    {
      code: `
  import { type Foo } from 'bar';
  
  declare module 'bar' {
	type Foo = string;
  }
		`,
    },
    {
      code: `
  import { type Foo } from 'bar';
  
  declare module 'bar' {
	interface Foo {
	  x: string;
	}
  }
		`,
    },

    // ---- Function-name initializer exception covers arbitrary wrappers
    // (ESLint scope-based check: outerScope === innerScope.upper).
    'const a = wrap(function a() {});',
    'const a = foo || wrap(function a() {});',
    'const { a = wrap(function a() {}) } = obj;',
    'const { a = foo || wrap(function a() {}) } = obj;',
    'const { a = foo, b = function a() {} } = {}',
    'const { A = Foo, B = class A {} } = {}',
    'function foo(a = wrap(function a() {})) {}',
    'function foo(a = foo || wrap(function a() {})) {}',
    'const A = wrap(class A {});',
    'const A = foo || wrap(class A {});',
    'const { A = wrap(class A {}) } = obj;',
    'const { A = foo || wrap(class A {}) } = obj;',
    'function foo(A = wrap(class A {})) {}',
    'function foo(A = foo || wrap(class A {})) {}',
    'var a = function a() {} ? foo : bar',
    'var A = class A {} ? foo : bar',
    {
      code: 'let x = false; export const a = wrap(function a() { if (!x) { x = true; a(); } });',
      options: [{ hoist: 'all' }] as any,
    },
  ],
  invalid: [
    // ---- Core JS shadow with line/column ----
    {
      code: `function a(x) { var b = function c() { var x = 'foo'; }; }`,
      errors: [{ messageId: 'noShadow', line: 1, column: 44 }],
    },
    {
      code: `var a = (x) => { var b = () => { var x = 'foo'; }; }`,
      errors: [{ messageId: 'noShadow', line: 1, column: 38 }],
    },
    {
      code: `function a(x) { var b = function () { var x = 'foo'; }; }`,
      errors: [{ messageId: 'noShadow', line: 1, column: 43 }],
    },
    {
      code: `var x = 1; function a(x) { return ++x; }`,
      errors: [{ messageId: 'noShadow', line: 1, column: 23 }],
    },
    {
      code: 'var a=3; function b() { var a=10; }',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'var a=3; function b() { var a=10; }; setTimeout(function() { b(); }, 0);',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'var a=3; function b() { var a=10; var b=0; }; setTimeout(function() { b(); }, 0);',
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },

    { code: 'var x = 1; { let x = 2; }', errors: [{ messageId: 'noShadow' }] },
    {
      code: 'let x = 1; { const x = 2; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Default hoist ("functions") ----
    { code: '{ let a; } function a() {}', errors: [{ messageId: 'noShadow' }] },
    {
      code: '{ const a = 0; } function a() {}',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo() { let a; } function a() {}',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo() { var a; } function a() {}',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo(a) { } function a() {}',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- hoist: "all" — all permutations ----
    {
      code: '{ let a; } let a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '{ let a; } var a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '{ let a; } function a() {}',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '{ const a = 0; } const a = 1;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '{ const a = 0; } var a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '{ const a = 0; } function a() {}',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo() { let a; } let a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo() { let a; } var a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo() { let a; } function a() {}',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo() { var a; } let a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo() { var a; } var a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo() { var a; } function a() {}',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo(a) { } let a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo(a) { } var a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo(a) { } function a() {}',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- fn/class expression self-name ----
    {
      code: '(function a() { function a(){} })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '(function a() { class a{} })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '(function a() { (function a(){}); })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '(function a() { (class a{}); })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '(function() { var a = function(a) {}; })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '(function() { var a = function() { function a() {} }; })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '(function() { var a = function() { class a{} }; })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '(function() { var a = function() { (function a() {}); }; })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '(function() { var a = function() { (class a{}); }; })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '(function() { var a = class { constructor() { class a {} } }; })()',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'class A { constructor() { var A; } }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Nested multi-error ----
    {
      code: '(function a() { function a(){ function a(){} } })()',
      errors: [
        { messageId: 'noShadow', line: 1, column: 26 },
        { messageId: 'noShadow', line: 1, column: 40 },
      ],
    },

    // ---- builtinGlobals ----
    {
      code: 'function foo() { var Object = 0; }',
      options: [{ builtinGlobals: true }] as any,
      errors: [{ messageId: 'noShadowGlobal' }],
    },
    {
      code: 'var Object = 0;',
      options: [{ builtinGlobals: true }] as any,
      errors: [{ messageId: 'noShadowGlobal' }],
    },
    {
      code: '(function Array() {})',
      options: [{ builtinGlobals: true }] as any,
      errors: [{ messageId: 'noShadowGlobal', line: 1, column: 11 }],
    },

    // ---- allow mismatch ----
    {
      code: 'function foo(cb) { (function (cb) { cb(42); })(cb); }',
      errors: [{ messageId: 'noShadow', line: 1, column: 31 }],
    },

    // ---- Class static blocks ----
    {
      code: 'class C { static { let a; { let a; } } }',
      errors: [{ messageId: 'noShadow', line: 1, column: 33 }],
    },
    {
      code: 'class C { static { var C; } }',
      errors: [{ messageId: 'noShadow', line: 1, column: 24 }],
    },
    {
      code: 'class C { static { let C; } }',
      errors: [{ messageId: 'noShadow', line: 1, column: 24 }],
    },
    {
      code: 'var a; class C { static { var a; } }',
      errors: [{ messageId: 'noShadow', line: 1, column: 31 }],
    },
    {
      code: 'class C { static { var a; } } var a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 24 }],
    },
    {
      code: 'class C { static { let a; } } let a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 24 }],
    },
    {
      code: 'class C { static { var a; } } let a;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 24 }],
    },
    {
      code: 'class C { static { var a; class D { static { var a; } } } }',
      errors: [{ messageId: 'noShadow', line: 1, column: 50 }],
    },
    {
      code: 'class C { static { let a; class D { static { let a; } } } }',
      errors: [{ messageId: 'noShadow', line: 1, column: 50 }],
    },

    // ---- Hoist "all" with param list ----
    {
      code: 'let x = foo((x,y) => {});\nlet y;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },
    {
      code: 'let x = ((x,y) => {})();\nlet y;',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },

    // ---- ignoreOnInitialization mismatches (still reported) ----
    {
      code: 'const a = fn(()=>{ class C { fn () { const a = 42; return a } } return new C() })',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 44 }],
    },
    {
      code: 'function a() {}\nfoo(a => {});',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 2, column: 5 }],
    },
    {
      code: 'const a = fn(()=>{ function C() { this.fn=function() { const a = 42; return a } } return new C() });',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 62 }],
    },
    {
      code: 'const x = foo(() => { const bar = () => { return x => {}; }; return bar; });',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 50 }],
    },
    {
      code: 'const x = foo(() => { return { bar(x) {} }; });',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 36 }],
    },
    {
      code: 'const x = () => { foo(x => x); }',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 23 }],
    },
    {
      code: 'const foo = () => { let x; bar(x => x); }',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 32 }],
    },
    {
      code: 'foo(() => { const x = x => x; });',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 23 }],
    },
    {
      code: 'const foo = (x) => { bar(x => {}) }',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 26 }],
    },
    {
      code: 'const a = (()=>{ class C { fn () { const a = 42; return a } } return new C() })()',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 42 }],
    },
    {
      code: 'const x = () => { (x => x)(); }',
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow', line: 1, column: 20 }],
    },

    // (Removed: call-wrap and conditional-test-position cases — ESLint's
    // `outerScope === innerScope.upper` filter exempts them; corresponding
    // valid cases live in the valid section.)
    {
      code: '(function Array() {})',
      options: [{ builtinGlobals: true }] as any,
      errors: [{ messageId: 'noShadowGlobal', line: 1, column: 11 }],
    },
    {
      code: 'let a; { let b = (function a() {}) }',
      errors: [{ messageId: 'noShadow', line: 1, column: 28 }],
    },
    {
      code: 'let a = foo; { let b = (function a() {}) }',
      errors: [{ messageId: 'noShadow', line: 1, column: 34 }],
    },

    // ==============================================================
    // TypeScript invalid
    // ==============================================================

    {
      code: '\n  type T = 1;\n  {\n\ttype T = 2;\n  }\n\t\t',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  type T = 1;\n  function foo<T>(arg: T) {}\n\t\t',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  function foo<T>() {\n\treturn function <T>() {};\n  }\n\t\t',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  type T = string;\n  function foo<T extends (arg: any) => void>(arg: T) {}\n\t\t',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  const x = 1;\n  {\n\ttype x = string;\n  }\n\t\t',
      options: [{ ignoreTypeValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- ignoreFunctionTypeParameterNameValueShadow: false ----
    {
      code: '\n  const test = 1;\n  type Fn = (test: string) => typeof test;\n\t\t',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  const arg = 0;\n  interface Test {\n\t(arg: string): typeof arg;\n  }\n\t\t',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  const arg = 0;\n  interface Test {\n\tp1(arg: string): typeof arg;\n  }\n\t\t',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  const arg = 0;\n  declare function test(arg: string): typeof arg;\n\t\t',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  const arg = 0;\n  declare const test: (arg: string) => typeof arg;\n\t\t',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  const arg = 0;\n  declare class Test {\n\tp1(arg: string): typeof arg;\n  }\n\t\t',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  const arg = 0;\n  declare const Test: {\n\tnew (arg: string): typeof arg;\n  };\n\t\t',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  const arg = 0;\n  type Bar = new (arg: number) => typeof arg;\n\t\t',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  const arg = 0;\n  declare namespace Lib {\n\tfunction test(arg: string): typeof arg;\n  }\n\t\t',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Type import with ignoreTypeValueShadow: false ----
    {
      code: "\nimport type { foo } from './foo';\nfunction doThing(foo: number) {}\n",
      options: [{ ignoreTypeValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: "\nimport { type foo } from './foo';\nfunction doThing(foo: number) {}\n",
      options: [{ ignoreTypeValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: "\nimport { foo } from './foo';\nfunction doThing(foo: number, bar: number) {}\n",
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Module augmentation interface shadow ----
    {
      code: "\ninterface Foo {}\ndeclare module 'bar' { export interface Foo { x: string } }\n",
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: "\nimport type { Foo } from 'bar';\ndeclare module 'baz' { export interface Foo { x: string } }\n",
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: "\nimport { type Foo } from 'bar';\ndeclare module 'baz' { export interface Foo { x: string } }\n",
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- hoist: all with TS ----
    {
      code: '\n  let x = foo((x, y) => {});\n  let y;\n\t\t',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },
    {
      code: '\n  let x = foo((x, y) => {});\n  let y;\n\t\t',
      options: [{ hoist: 'functions' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- TS hoist: types / functions-and-types / all ----
    {
      code: '\n  type Foo<A> = 1;\n  type A = 1;\n\t\t',
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  interface Foo<A> {}\n  type A = 1;\n\t\t',
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  interface Foo<A> {}\n  interface A {}\n\t\t',
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  type Foo<A> = 1;\n  interface A {}\n\t\t',
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  {\n\ttype A = 1;\n  }\n  type A = 1;\n\t\t',
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  {\n\tinterface A {}\n  }\n  type A = 1;\n\t\t',
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  type Foo<A> = 1;\n  type A = 1;\n\t\t',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  interface Foo<A> {}\n  type A = 1;\n\t\t',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  interface Foo<A> {}\n  interface A {}\n\t\t',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  type Foo<A> = 1;\n  interface A {}\n\t\t',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  {\n\ttype A = 1;\n  }\n  type A = 1;\n\t\t',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  {\n\tinterface A {}\n  }\n  type A = 1;\n\t\t',
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  type Foo<A> = 1;\n  type A = 1;\n\t\t',
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n\tif (true) {\n\t\tconst foo = 6;\n\t}\n\n\tfunction foo() { }\n\t\t',
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n\t// types\n\ttype Bar<Foo> = 1;\n\ttype Foo = 1;\n\n\t// functions\n\tif (true) {\n\t\tconst b = 6;\n\t}\n\n\tfunction b() { }\n\t\t',
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },
    {
      code: '\n\t// types\n\ttype Bar<Foo> = 1;\n\ttype Foo = 1;\n\n\t// functions\n\tif (true) {\n\t\tconst b = 6;\n\t}\n\n\tfunction b() { }\n\t\t',
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  interface Foo<A> {}\n  type A = 1;\n\t\t',
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  interface Foo<A> {}\n  interface A {}\n\t\t',
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  type Foo<A> = 1;\n  interface A {}\n\t\t',
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  {\n\ttype A = 1;\n  }\n  type A = 1;\n\t\t',
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: '\n  {\n\tinterface A {}\n  }\n  type A = 1;\n\t\t',
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Enum member shadow ----
    {
      code: '\n\t\tconst A = 2;\n\t\tenum Test {\n\t\t\tA = 1,\n\t\t\tB = A,\n\t\t}\n\t',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- infer T shadow ----
    {
      code: 'type X<T> = T extends (infer U) ? (U extends (infer U) ? U : never) : never;',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Object literal method param shadow ----
    {
      code: 'const x = 1; const o = { foo(x: number) { return x; } };',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'const x = 1; const o = { get foo() { const x = 2; return x; } };',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'const x = 1; const o = { async foo(x: number) { return x; } };',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'const x = 1; const o = { *foo(x: number) { yield x; } };',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- using declarations ----
    {
      code: 'using u = { [Symbol.dispose]() {} }; { using u = { [Symbol.dispose]() {} }; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Class generic shadowed by instance method generic ----
    {
      code: 'class C<T> { static s<T>(): T { return null as any; } i<T>(): T { return null as any; } }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Object pattern rename shadow ----
    {
      code: 'const a = 1; function f({ x: a }: any) { return a; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Function expression name shadowed by its own param ----
    {
      code: 'const fn = function f(f: number) { return f; };',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Decorator: shadowing inside decorated method body ----
    {
      code: 'function dec(x: any) { return x; }\n@dec class CD { method() { const dec = 1; return dec; } }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Computed property name with class member shadow ----
    {
      code: "const k = 'x'; class C { foo() { const k = 1; return k; } }",
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Default param referencing outer var of same name ----
    {
      code: 'const a = 1; function f({ a = a }: any) { return a; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Catch-clause destructure shadow ----
    {
      code: 'const e = 1; try {} catch ({ message: e }) { return e; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Async generator with for-await-of ----
    {
      code: 'async function* g() { const v = 1; for await (const v of [Promise.resolve(1)]) yield v; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Multi-decl single statement ----
    {
      code: 'const a = 1, b = 2; function f() { const a = 3, b = 4; return a + b; }',
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },

    // ---- Switch case lexical scope ----
    {
      code: 'function f() { const z = 1; switch (z) { case 1: { const z = 2; break; } case 2: { const z = 3; break; } } }',
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },

    // ---- Namespace import param shadow ----
    {
      code: "import * as Mods from 'fs'; function f(Mods: number) { return Mods; }",
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Arrow returning class with shadowed name ----
    {
      code: 'const Z = 1; const make = () => class Z {};',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- typeof reference + parameter shadow ----
    {
      code: 'const t = 1; type T = typeof t; function f(t: number) { return t; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Default parameter with type reference ----
    {
      code: 'const opts = { x: 1 }; function f(opts: typeof opts = opts) { return opts; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Anonymous function expression with same-name inner const ----
    {
      code: 'const a = 1; const fn = function() { const a = 2; return a; };',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Nested class with shadowed name ----
    {
      code: 'const A = 1; class A_outer { x = class A {}; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Rest pattern shadow inner destructure ----
    {
      code: 'function f(...args: number[]) { function g({ args }: any) { return args; } }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Generator with for-of shadow ----
    {
      code: 'function* g() { const v = 1; for (const v of [1,2,3]) { yield v; } }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Function type in parameter annotation with shadow generic ----
    {
      code: 'function f<A>(fn: <A>(x: A) => A): A { return fn as any; }',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function f<T>(): <T>(x: T) => T { return null as any; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Nested catch clause ----
    {
      code: 'try {} catch (e) { try {} catch (e) {} }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Class generic shadowed by method generic ----
    {
      code: 'class C<T> { m<T>(x: T): T { return x; } }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Double shadow chain ----
    {
      code: 'const x = 1; function f(x: number) { function g(x: number) { return x; } return g; }',
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },

    // ---- Array elision pattern ----
    {
      code: 'const a = 1; function f([, a]: any[]) { return a; }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Optional chain call with arrow ----
    {
      code: 'const a = 1; const fn = { call: (cb: any) => cb }; fn.call?.(a => a);',
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- Parameter property + method shadow ----
    {
      code: 'const x = 1; class C { constructor(public x: number) {} m() { const x = 1; return x; } }',
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },

    // ---- Dynamic import callback ----
    {
      code: "const mod = 1; import('m').then((mod) => mod);",
      errors: [{ messageId: 'noShadow' }],
    },

    // ---- TS 4.7 infer T extends + nested same name ----
    {
      code: 'type X<T> = T extends Array<infer U> ? (U extends Array<infer U> ? U : never) : never;',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- TS 4.7 generic instantiation expression shadowing builtin Array ----
    {
      code: 'const g = Array<number>; function f(Array: any) { return Array; }',
      options: [{ builtinGlobals: true }] as any,
      errors: [{ messageId: 'noShadowGlobal' }],
    },
    // ---- TS 4.9 accessor ----
    {
      code: 'const x = 1; class C { accessor x = 1; m() { const x = 2; return x; } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Async iteration destructure shadow ----
    {
      code: 'async function f() { const x = 1; for await (const { v: x } of [Promise.resolve({ v: 1 })]) { return x; } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- using + for-of shadow ----
    {
      code: 'async function f() { using u = { [Symbol.dispose]() {} }; for (const u of [] as any[]) { void u; } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Destructure with computed key + inner rebind ----
    {
      code: "const k = 'a'; function f({ [k]: v }: any) { const k = 1; return [v, k]; }",
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- HOC pattern ----
    {
      code: 'function withTheme<P>(Component: any) { return function ThemedComponent(props: P) { const Component = 1; return Component; }; }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Reducer with case block ----
    {
      code: "type A = { type: 'inc' } | { type: 'set'; value: number }; function reducer(state: number, action: A) { switch (action.type) { case 'inc': { const state = 1; return state + 1; } case 'set': return action.value; } }",
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Array chain with 3 same-name callbacks ----
    {
      code: 'function chain() { const item = { id: 1 }; [1, 2, 3].map(item => item + 1).filter(item => item > 1).forEach(item => void item); void item; }',
      errors: [
        { messageId: 'noShadow' },
        { messageId: 'noShadow' },
        { messageId: 'noShadow' },
      ],
    },
    // ---- Triply nested try/catch ----
    {
      code: 'const e = 1; try { try { throw new Error(); } catch (e) { try { throw e; } catch (e) {} } } catch (e) {} void e;',
      errors: [
        { messageId: 'noShadow' },
        { messageId: 'noShadow' },
        { messageId: 'noShadow' },
      ],
    },
    // ---- Computed method name + param shadow ----
    {
      code: "const methodName = 'foo'; const obj = { [methodName](methodName: string) { return methodName; } };",
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Nested for loops with same loop var ----
    {
      code: 'function f() { for (let i = 0; i < 1; i++) { for (let i = 0; i < 1; i++) {} } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Method overload: impl-signature param shadow ----
    {
      code: 'const a = 1; class C { m(x: string): void; m(x: number): void; m(a: any): void { void a; } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Factory returning obj method with generic shadow ----
    {
      code: 'function mk<T>(x: T) { return { get<T>(): T { return null as any; } }; }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Empty block followed by shadowed block ----
    {
      code: 'const c = 1; { /* empty */ } { const c = 2; void c; }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- const enum + param shadow ----
    {
      code: 'const enum E { A, B } function f(E: number) { return E; }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Abstract class generic shadow ----
    {
      code: 'abstract class C<T extends object> { abstract m<T>(): T; }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Symbol.iterator method with shadowed param ----
    {
      code: 'const iter = 1; class C { [Symbol.iterator](iter: number) { return iter; } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Constructor with optional param shadow ----
    {
      code: 'const x = 1; class C { constructor(x?: number) { void x; } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- typeof type query + same-name parameter ----
    {
      code: 'const val = { x: 1 }; function f(val: typeof val) { return val; }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Async method param shadow ----
    {
      code: 'const x = 1; class C { async m(x: number) { return x; } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Generator method param shadow ----
    {
      code: 'const x = 1; class C { *m(x: number) { yield x; } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Ambient function shadowed by inner const (bot #1 / #3) ----
    {
      code: 'declare function foo(): void; function bar() { const foo = 1; return foo; }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Heritage clause IIFE shadow (bot #2) ----
    {
      code: 'const h = 1; class C extends (function() { const h = 2; return class {}; })() {}',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Heritage clause comma-expression arrow shadow ----
    {
      code: 'const h = 1; class C extends ((h => h)(1), Object) {}',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Class decorator factory body shadow (bot #2) ----
    {
      code: 'function dec(x: any) { return x; }\n@((target: any) => { const dec = 1; return dec && target; })\nclass D {}',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Method decorator body shadow ----
    {
      code: 'const md = 1; class C { @((t: any, k: string) => { const md = 1; void md; }) method() {} }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Parameter decorator body shadow ----
    {
      code: 'const pd = 1; class C { method(@((t: any, k: any, i: number) => { const pd = 1; void pd; }) x: number) {} }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Computed key on class method (self-discovered during audit) ----
    {
      code: 'const k = 1; class C { [((k) => k)(2)]() {} }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Computed key on class property ----
    {
      code: 'const k = 1; class C { [((k) => k)(2)] = 1; }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Computed key on getter ----
    {
      code: 'const g = 1; class C { get [((g) => g)(1)]() { return 1; } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // ---- Computed key on async generator method ----
    {
      code: 'const m = 1; class C { async *[((m) => m)(1)]() {} }',
      errors: [{ messageId: 'noShadow' }],
    },

    // ==== Additional invalid cases mirroring Go ====
    {
      code: `
  {
	interface A {}
  }
  type A = 1;
		`,
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  {
	interface A {}
  }
  type A = 1;
		`,
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
	if (true) {
		const foo = 6;
	}

	function foo() { }
		`,
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
	// types
	type Bar<Foo> = 1;
	type Foo = 1;

	// functions
	if (true) {
		const b = 6;
	}

	function b() { }
		`,
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },
    {
      code: `
	// types
	type Bar<Foo> = 1;
	type Foo = 1;

	// functions
	if (true) {
		const b = 6;
	}

	function b() { }
		`,
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  {
	interface A {}
  }
  type A = 1;
		`,
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `let x = foo((x,y) => {});
let y;`,
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },
    {
      code: `function a() {}
foo(a => {});`,
      options: [{ ignoreOnInitialization: true }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `let x = ((x,y) => {})();
let y;`,
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },
    {
      code: `
  type T = 1;
  {
	type T = 2;
  }
		`,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  type T = 1;
  function foo<T>(arg: T) {}
		`,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  function foo<T>() {
	return function <T>() {};
  }
		`,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  type T = string;
  function foo<T extends (arg: any) => void>(arg: T) {}
		`,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  const arg = 0;
  
  declare const Test: {
	new (arg: string): typeof arg;
  };
		`,
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  import type { foo } from './foo';
  function doThing(foo: number) {}
		`,
      options: [{ ignoreTypeValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  import { type foo } from './foo';
  function doThing(foo: number) {}
		`,
      options: [{ ignoreTypeValueShadow: false }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  import { foo } from './foo';
  function doThing(foo: number, bar: number) {}
		`,
      options: [{ ignoreTypeValueShadow: true }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  interface Foo {}
  
  declare module 'bar' {
	export interface Foo {
	  x: string;
	}
  }
		`,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  import type { Foo } from 'bar';
  
  declare module 'baz' {
	export interface Foo {
	  x: string;
	}
  }
		`,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  import { type Foo } from 'bar';
  
  declare module 'baz' {
	export interface Foo {
	  x: string;
	}
  }
		`,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  let x = foo((x, y) => {});
  let y;
		`,
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },
    {
      code: `
  let x = foo((x, y) => {});
  let y;
		`,
      options: [{ hoist: 'functions' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  {
	interface A {}
  }
  type A = 1;
		`,
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  {
	interface A {}
  }
  type A = 1;
		`,
      options: [{ hoist: 'all' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
	if (true) {
		const foo = 6;
	}

	function foo() { }
		`,
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
	// types
	type Bar<Foo> = 1;
	type Foo = 1;

	// functions
	if (true) {
		const b = 6;
	}

	function b() { }
		`,
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }, { messageId: 'noShadow' }],
    },
    {
      code: `
	// types
	type Bar<Foo> = 1;
	type Foo = 1;

	// functions
	if (true) {
		const b = 6;
	}

	function b() { }
		`,
      options: [{ hoist: 'types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
  {
	interface A {}
  }
  type A = 1;
		`,
      options: [{ hoist: 'functions-and-types' }] as any,
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: `
			const A = 2;
			enum Test {
				A = 1,
				B = A,
			}
		`,
      errors: [{ messageId: 'noShadow' }],
    },
  ],
});
