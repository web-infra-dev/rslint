import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();
const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('no-shadow', {
  valid: [
    // Basic cases
    {
      code: 'var a = 3; function b() { var c = a; }',
    },
    {
      code: 'function a() {} function b() { var a = 10; }',
      options: [{ hoist: 'never' }],
    },
    {
      code: 'var a = 3; function b() { var a = 10; }',
      options: [{ allow: ['a'] }],
    },
    {
      code: 'function foo() { var Object = 0; }',
      options: [{ builtinGlobals: false }],
    },
    // TypeScript specific - type/value shadowing
    {
      code: 'type Foo = string; function test() { const Foo = 1; }',
      options: [{ ignoreTypeValueShadow: true }],
    },
    {
      code: 'interface Foo {} class Bar { Foo: string; }',
      options: [{ ignoreTypeValueShadow: true }],
    },
    {
      code: 'type Foo = number; interface Bar { prop: number; } function f() { const Foo = 1; const Bar = "test"; }',
      options: [{ ignoreTypeValueShadow: true }],
    },
    // Function type parameter shadowing
    {
      code: 'let test = 1; type TestType = typeof test; type Func = (test: string) => typeof test;',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: true }],
    },
    {
      code: 'type Fn = (Foo: string) => void; const Foo = "bar";',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: true }],
    },
    // Enums
    {
      code: 'const test = "test"; export enum MyEnum { test = 42 }',
      options: [{ ignoreTypeValueShadow: true }],
    },
    {
      code: 'enum Foo { a = 1, b = 2 } function bar() { const a = "test"; }',
      options: [{ ignoreTypeValueShadow: true }],
    },
    // Import declarations
    {
      code: 'import type { Foo } from "./foo"; function bar(Foo: string) {}',
      options: [{ ignoreTypeValueShadow: true }],
    },
    // Class and function expressions
    {
      code: 'var a = function a() { return a; };',
    },
    {
      code: 'class Foo { bar() { class Foo {} } }',
      options: [{ allow: ['Foo'] }],
    },
    // Hoisting
    {
      code: 'function a() { return b; function b() {} }',
      options: [{ hoist: 'functions' }],
    },
    {
      code: 'function a() { return b; interface b {} }',
      options: [{ hoist: 'types' }],
    },
    {
      code: 'function a() { return b; function b() {} interface c {} }',
      options: [{ hoist: 'functions-and-types' }],
    },
    // Module declarations
    {
      code: 'declare module "foo" { export interface Bar {} } const Bar = 1;',
      options: [{ ignoreTypeValueShadow: true }],
    },
    // Static method generics
    {
      code: 'class Foo<T> { static method<T>() {} }',
    },
    // This parameters
    {
      code: 'function foo(this: any, a: number) { function bar(this: any, a: string) {} }',
    },
  ],
  invalid: [
    // Basic shadowing
    {
      code: 'var a = 3; function b() { var a = 10; }',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'function foo() { var Object = 0; }',
      options: [{ builtinGlobals: true }],
      errors: [{ messageId: 'noShadowGlobal' }],
    },
    {
      code: 'var a = 3; var a = 10;',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'var a = 3; function b() { function a() {} }',
      options: [{ hoist: 'never' }],
      errors: [{ messageId: 'noShadow' }],
    },
    // TypeScript specific - when options are disabled
    {
      code: 'type Foo = string; function test() { const Foo = 1; }',
      options: [{ ignoreTypeValueShadow: false }],
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'let test = 1; type Func = (test: string) => typeof test;',
      options: [{ ignoreFunctionTypeParameterNameValueShadow: false }],
      errors: [{ messageId: 'noShadow' }],
    },
    // Nested function parameters
    {
      code: 'function foo(a: number) { function bar() { function baz(a: string) {} } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // Class shadowing
    {
      code: 'class A { method() { class A {} } }',
      errors: [{ messageId: 'noShadow' }],
    },
    // Block scope shadowing
    {
      code: 'let x = 1; { let x = 2; }',
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'const a = 1; for (const a of [1, 2, 3]) {}',
      errors: [{ messageId: 'noShadow' }],
    },
    // Catch clause
    {
      code: 'const e = 1; try {} catch (e) {}',
      errors: [{ messageId: 'noShadow' }],
    },
    // Module and enum shadowing
    {
      code: 'const Foo = 1; enum Foo { Bar }',
      options: [{ ignoreTypeValueShadow: false }],
      errors: [{ messageId: 'noShadow' }],
    },
    {
      code: 'const Foo = 1; namespace Foo {}',
      options: [{ ignoreTypeValueShadow: false }],
      errors: [{ messageId: 'noShadow' }],
    },
    // Arrow functions
    {
      code: 'const a = 1; const fn = (a: number) => a;',
      errors: [{ messageId: 'noShadow' }],
    },
    // Method parameters
    {
      code: 'const a = 1; class C { method(a: number) {} }',
      errors: [{ messageId: 'noShadow' }],
    },
    // Constructor parameters
    {
      code: 'const a = 1; class C { constructor(a: number) {} }',
      errors: [{ messageId: 'noShadow' }],
    },
    // Getter/setter parameters
    {
      code: 'const a = 1; class C { set prop(a: number) {} }',
      errors: [{ messageId: 'noShadow' }],
    },
    // Multiple shadowing in different scopes
    {
      code: 'var x = 1; function a() { var x = 2; function b() { var x = 3; } }',
      errors: [
        { messageId: 'noShadow', data: { name: 'x' } },
        { messageId: 'noShadow', data: { name: 'x' } },
      ],
    },
  ],
});
