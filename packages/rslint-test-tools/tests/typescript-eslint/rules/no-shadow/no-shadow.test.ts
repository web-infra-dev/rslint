import { RuleTester } from '@rslint/rule-tester';
import { config } from '../../config';

// Preserve the existing Go binding ranges and enum-shadow diagnostic while strengthening the assertions.
new RuleTester(config).run('no-shadow', {
  valid: [
    {
      code: 'function fn() { type Record = string; }',
      options: [
        {
          builtinGlobals: true,
        },
      ],
    },
    {
      code: 'function fn(Record: number) {}',
      options: [
        {
          builtinGlobals: false,
        },
      ],
    },
    {
      code: 'function fn(NodeListOf: number, HTMLElement: number) {}',
      options: [
        {
          builtinGlobals: true,
        },
      ],
    },
    'function foo<T = (arg: any) => any>(arg: T) {}',
    'function foo<T = ([arg]: [any]) => any>(arg: T) {}',
    'function foo<T = ({ args }: { args: any }) => any>(arg: T) {}',
    'function foo<T = (...args: any[]) => any>(fn: T, args: any[]) {}',
    'function foo<T extends (...args: any[]) => any>(fn: T, args: any[]) {}',
    'function foo<T extends (...args: any[]) => any>(fn: T, ...args: any[]) {}',
    'function foo<T extends ([args]: any[]) => any>(fn: T, args: any[]) {}',
    'function foo<T extends ([...args]: any[]) => any>(fn: T, args: any[]) {}',
    'function foo<T extends ({ args }: { args: any }) => any>(fn: T, args: any) {}',
    `
function foo<T extends (id: string, ...args: any[]) => any>(
  fn: T,
  ...args: any[]
) {}
    `,
    `
type Args = 1;
function foo<T extends (Args: any) => void>(arg: T) {}
    `,
    // nested conditional types
    `
export type ArrayInput<Func> = Func extends (arg0: Array<infer T>) => any
  ? T[]
  : Func extends (...args: infer T) => any
    ? T
    : never;
    `,
    `
function foo() {
  var Object = 0;
}
    `,
    // this params
    `
function test(this: Foo) {
  function test2(this: Bar) {}
}
    `,
    // declaration merging
    `
class Foo {
  prop = 1;
}
namespace Foo {
  export const v = 2;
}
    `,
    `
function Foo() {}
namespace Foo {
  export const v = 2;
}
    `,
    `
class Foo {
  prop = 1;
}
interface Foo {
  prop2: string;
}
    `,
    `
import type { Foo } from 'bar';

declare module 'bar' {
  export interface Foo {
    x: string;
  }
}
    `,
    // type value shadowing
    `
const x = 1;
type x = string;
    `,
    `
const x = 1;
{
  type x = string;
}
    `,
    {
      code: `
type Foo = 1;
      `,
      options: [
        {
          ignoreTypeValueShadow: true,
        },
      ],
      languageOptions: {
        globals: {
          Foo: 'writable',
        },
      },
    },
    {
      code: `
type Foo = 1;
      `,
      options: [
        {
          builtinGlobals: false,
          ignoreTypeValueShadow: false,
        },
      ],
      languageOptions: {
        globals: {
          Foo: 'writable',
        },
      },
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/2360
    `
enum Direction {
  left = 'left',
  right = 'right',
}
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/2447
    {
      code: `
const test = 1;
type Fn = (test: string) => typeof test;
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: `
type Fn = (Foo: string) => typeof Foo;
      `,
      options: [
        {
          builtinGlobals: false,
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
      languageOptions: {
        globals: {
          Foo: 'writable',
        },
      },
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/6098
    {
      code: `
const arg = 0;

interface Test {
  (arg: string): typeof arg;
}
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: `
const arg = 0;

interface Test {
  p1(arg: string): typeof arg;
}
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare function test(arg: string): typeof arg;
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare const test: (arg: string) => typeof arg;
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare class Test {
  p1(arg: string): typeof arg;
}
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare const Test: {
  new (arg: string): typeof arg;
};
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: `
const arg = 0;

type Bar = new (arg: number) => typeof arg;
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare namespace Lib {
  function test(arg: string): typeof arg;
}
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: true,
        },
      ],
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/2724
    {
      code: `
        declare global {
          interface ArrayConstructor {}
        }
        export {};
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
    },
    `
      declare global {
        const a: string;

        namespace Foo {
          const a: number;
        }
      }
      export {};
    `,
    {
      code: `
        declare global {
          type A = 'foo';

          namespace Foo {
            type A = 'bar';
          }
        }
        export {};
      `,
      options: [
        {
          ignoreTypeValueShadow: false,
        },
      ],
    },
    {
      code: `
        declare global {
          const foo: string;
          type Fn = (foo: number) => void;
        }
        export {};
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
    },
    `
export class Wrapper<Wrapped> {
  private constructor(private readonly wrapped: Wrapped) {}

  unwrap(): Wrapped {
    return this.wrapped;
  }

  static create<Wrapped>(wrapped: Wrapped) {
    return new Wrapper<Wrapped>(wrapped);
  }
}
    `,
    `
function makeA() {
  return class A<T> {
    constructor(public value: T) {}

    static make<T>(value: T) {
      return new A<T>(value);
    }
  };
}
    `,
    // https://github.com/typescript-eslint/typescript-eslint/issues/3862
    {
      code: `
import type { foo } from './foo';
type bar = number;

// 'foo' is already declared in the upper scope
// 'bar' is fine
function doThing(foo: number, bar: number) {}
      `,
      options: [
        {
          ignoreTypeValueShadow: true,
        },
      ],
    },
    {
      code: `
import { type foo } from './foo';

// 'foo' is already declared in the upper scope
function doThing(foo: number) {}
      `,
      options: [
        {
          ignoreTypeValueShadow: true,
        },
      ],
    },
    {
      code: 'const a = [].find(a => a);',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: `
const a = [].find(function (a) {
  return a;
});
      `,
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'const [a = [].find(a => true)] = dummy;',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'const { a = [].find(a => true) } = dummy;',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'function func(a = [].find(a => true)) {}',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: `
for (const a in [].find(a => true)) {
}
      `,
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: `
for (const a of [].find(a => true)) {
}
      `,
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const a = [].map(a => true).filter(a => a === 'b');",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: `
const a = []
  .map(a => true)
  .filter(a => a === 'b')
  .find(a => a === 'c');
      `,
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'const { a } = (({ a }) => ({ a }))();',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: `
const person = people.find(item => {
  const person = item.name;
  return person === 'foo';
});
      `,
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'var y = bar || foo(y => y);',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'var y = bar && foo(y => y);',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'var z = bar(foo(z => z));',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'var z = boo(bar(foo(z => z)));',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: `
var match = function (person) {
  return person.name === 'foo';
};
const person = [].find(match);
      `,
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'const a = foo(x || (a => {}));',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'const { a = 1 } = foo(a => {});',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: "const person = { ...people.find(person => person.firstName.startsWith('s')) };",
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2021,
      },
    },
    {
      code: `
const person = {
  firstName: people
    .filter(person => person.firstName.startsWith('s'))
    .map(person => person.firstName)[0],
};
      `,
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2021,
      },
    },
    {
      code: `
() => {
  const y = foo(y => y);
};
      `,
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'const x = (x => x)();',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'var y = bar || (y => y)();',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'var y = bar && (y => y)();',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'var x = (x => x)((y => y)());',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: 'const { a = 1 } = (a => {})();',
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    {
      code: `
() => {
  const y = (y => y)();
};
      `,
      options: [
        {
          ignoreOnInitialization: true,
        },
      ],
    },
    'const [x = y => y] = [].map(y => y);',
    {
      code: `
type Foo<A> = 1;
type A = 1;
      `,
      options: [
        {
          hoist: 'never',
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
type A = 1;
      `,
      options: [
        {
          hoist: 'never',
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
interface A {}
      `,
      options: [
        {
          hoist: 'never',
        },
      ],
    },
    {
      code: `
type Foo<A> = 1;
interface A {}
      `,
      options: [
        {
          hoist: 'never',
        },
      ],
    },
    {
      code: `
{
  type A = 1;
}
type A = 1;
      `,
      options: [
        {
          hoist: 'never',
        },
      ],
    },
    {
      code: `
{
  interface Foo<A> {}
}
type A = 1;
      `,
      options: [
        {
          hoist: 'never',
        },
      ],
    },
    {
      code: `
type Foo<A> = 1;
type A = 1;
      `,
      options: [
        {
          hoist: 'functions',
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
type A = 1;
      `,
      options: [
        {
          hoist: 'functions',
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
interface A {}
      `,
      options: [
        {
          hoist: 'functions',
        },
      ],
    },
    {
      code: `
type Foo<A> = 1;
interface A {}
      `,
      options: [
        {
          hoist: 'functions',
        },
      ],
    },
    {
      code: `
{
  type A = 1;
}
type A = 1;
      `,
      options: [
        {
          hoist: 'functions',
        },
      ],
    },
    {
      code: `
{
  interface Foo<A> {}
}
type A = 1;
      `,
      options: [
        {
          hoist: 'functions',
        },
      ],
    },
    `
import type { Foo } from 'bar';

declare module 'bar' {
  export type Foo = string;
}
      `,
    `
import type { Foo } from 'bar';

declare module 'bar' {
  interface Foo {
    x: string;
  }
}
      `,
    `
import { type Foo } from 'bar';

declare module 'bar' {
  export type Foo = string;
}
      `,
    `
import { type Foo } from 'bar';

declare module 'bar' {
  export interface Foo {
    x: string;
  }
}
      `,
    `
import { type Foo } from 'bar';

declare module 'bar' {
  type Foo = string;
}
      `,
    `
import { type Foo } from 'bar';

declare module 'bar' {
  interface Foo {
    x: string;
  }
}
      `,
    {
      code: `
declare const foo1: boolean;
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo1: false,
        },
      },
      filename: 'decl.d.ts',
    },
    {
      code: `
declare let foo2: boolean;
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo2: false,
        },
      },
      filename: 'decl.d.ts',
    },
    {
      code: `
declare var foo3: boolean;
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo3: false,
        },
      },
      filename: 'decl.d.ts',
    },
    {
      code: `
function foo4(name: string): void;
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo4: false,
        },
      },
      filename: 'decl.d.ts',
    },
    {
      code: `
declare class Foopy1 {
  name: string;
}
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy1: false,
        },
      },
      filename: 'decl.d.ts',
    },
    {
      code: `
declare interface Foopy2 {
  name: string;
}
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy2: false,
        },
      },
      filename: 'decl.d.ts',
    },
    {
      code: `
declare type Foopy3 = {
  x: number;
};
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy3: false,
        },
      },
      filename: 'decl.d.ts',
    },
    {
      code: `
declare enum Foopy4 {
  x,
}
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy4: false,
        },
      },
      filename: 'decl.d.ts',
    },
    {
      code: `
declare namespace Foopy5 {}
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          Foopy5: false,
        },
      },
      filename: 'decl.d.ts',
    },
    {
      code: `
declare;
foo5: boolean;
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          foo5: false,
        },
      },
      filename: 'decl.d.ts',
    },
  ],
  invalid: [
    {
      code: 'const f = (function f() {} as unknown);',
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'f' is already declared in the upper scope on line 1 column 7.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'enum A { A }',
      errors: [
        {
          messageId: 'noEnumShadow',
          message:
            "Enum members are added to the enum scope, so references to 'A' in enum member initializers resolve to this member instead of the declaration in the upper scope on line 1 column 6.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'function fn(Record: number) {}',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'Record' is already a global variable.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'function fn() { type Record = string; }',
      options: [
        {
          builtinGlobals: true,
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'Record' is already a global variable.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: "import type { Record } from 'pkg';",
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'Record' is already a global variable.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'type Fn = (Record: number) => void;',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'Record' is already a global variable.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'interface Record {} namespace Record { export const value = 1; }',
      options: [
        {
          builtinGlobals: true,
        },
      ],
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'Record' is already a global variable.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'class C { static method<Record>() {} }',
      options: [
        {
          builtinGlobals: true,
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'Record' is already a global variable.",
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
import { request, type Server } from 'node:http';
function use(request: unknown) {}
      `,
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'request' is already declared in the upper scope on line 2 column 10.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
type T = 1;
{
  type T = 2;
}
      `,
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'T' is already declared in the upper scope on line 2 column 6.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
type T = 1;
function foo<T>(arg: T) {}
      `,
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'T' is already declared in the upper scope on line 2 column 6.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
function foo<T>() {
  return function <T>() {};
}
      `,
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'T' is already declared in the upper scope on line 2 column 14.",
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
type T = string;
function foo<T extends (arg: any) => void>(arg: T) {}
      `,
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'T' is already declared in the upper scope on line 2 column 6.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
const x = 1;
{
  type x = string;
}
      `,
      options: [
        {
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'x' is already declared in the upper scope on line 2 column 7.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
type Foo = 1;
      `,
      options: [
        {
          builtinGlobals: true,
          ignoreTypeValueShadow: false,
        },
      ],
      languageOptions: {
        globals: {
          Foo: 'writable',
        },
      },
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'Foo' is already a global variable.",
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
const test = 1;
type Fn = (test: string) => typeof test;
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'test' is already declared in the upper scope on line 2 column 7.",
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
type Fn = (Foo: string) => typeof Foo;
      `,
      options: [
        {
          builtinGlobals: true,
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      languageOptions: {
        globals: {
          Foo: 'writable',
        },
      },
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'Foo' is already a global variable.",
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
const arg = 0;

interface Test {
  (arg: string): typeof arg;
}
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 4,
          endLine: 5,
          endColumn: 7,
        },
      ],
    },
    {
      code: `
const arg = 0;

interface Test {
  p1(arg: string): typeof arg;
}
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 6,
          endLine: 5,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare function test(arg: string): typeof arg;
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 4,
          column: 23,
          endLine: 4,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare const test: (arg: string) => typeof arg;
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare class Test {
  p1(arg: string): typeof arg;
}
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 6,
          endLine: 5,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare const Test: {
  new (arg: string): typeof arg;
};
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 8,
          endLine: 5,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
const arg = 0;

type Bar = new (arg: number) => typeof arg;
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
const arg = 0;

declare namespace Lib {
  function test(arg: string): typeof arg;
}
      `,
      options: [
        {
          ignoreFunctionTypeParameterNameValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'arg' is already declared in the upper scope on line 2 column 7.",
          line: 5,
          column: 17,
          endLine: 5,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
import type { foo } from './foo';
function doThing(foo: number) {}
      `,
      options: [
        {
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'foo' is already declared in the upper scope on line 2 column 15.",
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
import { type foo } from './foo';
function doThing(foo: number) {}
      `,
      options: [
        {
          ignoreTypeValueShadow: false,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'foo' is already declared in the upper scope on line 2 column 15.",
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
import { foo } from './foo';
function doThing(foo: number, bar: number) {}
      `,
      options: [
        {
          ignoreTypeValueShadow: true,
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'foo' is already declared in the upper scope on line 2 column 10.",
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 21,
        },
      ],
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
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'Foo' is already declared in the upper scope on line 2 column 11.",
          line: 5,
          column: 20,
          endLine: 5,
          endColumn: 23,
        },
      ],
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
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'Foo' is already declared in the upper scope on line 2 column 15.",
          line: 5,
          column: 20,
          endLine: 5,
          endColumn: 23,
        },
      ],
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
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'Foo' is already declared in the upper scope on line 2 column 15.",
          line: 5,
          column: 20,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
let x = foo((x, y) => {});
let y;
      `,
      options: [
        {
          hoist: 'all',
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'x' is already declared in the upper scope on line 2 column 5.",
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 15,
        },
        {
          messageId: 'noShadow',
          message:
            "'y' is already declared in the upper scope on line 3 column 5.",
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: `
let x = foo((x, y) => {});
let y;
      `,
      options: [
        {
          hoist: 'functions',
        },
      ],
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'x' is already declared in the upper scope on line 2 column 5.",
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
type Foo<A> = 1;
type A = 1;
      `,
      options: [
        {
          hoist: 'types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
type A = 1;
      `,
      options: [
        {
          hoist: 'types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
interface A {}
      `,
      options: [
        {
          hoist: 'types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
type Foo<A> = 1;
interface A {}
      `,
      options: [
        {
          hoist: 'types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
{
  type A = 1;
}
type A = 1;
      `,
      options: [
        {
          hoist: 'types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
{
  interface A {}
}
type A = 1;
      `,
      options: [
        {
          hoist: 'types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
type Foo<A> = 1;
type A = 1;
      `,
      options: [
        {
          hoist: 'all',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
type A = 1;
      `,
      options: [
        {
          hoist: 'all',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
interface A {}
      `,
      options: [
        {
          hoist: 'all',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
type Foo<A> = 1;
interface A {}
      `,
      options: [
        {
          hoist: 'all',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
{
  type A = 1;
}
type A = 1;
      `,
      options: [
        {
          hoist: 'all',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
{
  interface A {}
}
type A = 1;
      `,
      options: [
        {
          hoist: 'all',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
type Foo<A> = 1;
type A = 1;
      `,
      options: [
        {
          hoist: 'functions-and-types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
type A = 1;
      `,
      options: [
        {
          hoist: 'functions-and-types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 6.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
interface Foo<A> {}
interface A {}
      `,
      options: [
        {
          hoist: 'functions-and-types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
type Foo<A> = 1;
interface A {}
      `,
      options: [
        {
          hoist: 'functions-and-types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 3 column 11.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: `
{
  type A = 1;
}
type A = 1;
      `,
      options: [
        {
          hoist: 'functions-and-types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
{
  interface A {}
}
type A = 1;
      `,
      options: [
        {
          hoist: 'functions-and-types',
        },
      ],
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'A' is already declared in the upper scope on line 5 column 6.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
function foo<T extends (...args: any[]) => any>(fn: T, args: any[]) {}
      `,
      options: [
        {
          builtinGlobals: true,
          ignoreTypeValueShadow: false,
        },
      ],
      languageOptions: {
        globals: {
          args: 'writable',
        },
      },
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'args' is already a global variable.",
          line: 2,
          column: 56,
          endLine: 2,
          endColumn: 60,
        },
      ],
    },
    {
      code: `
declare const has = (environment: 'dev' | 'prod' | 'test') => boolean;
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          has: false,
        },
      },
      errors: [
        {
          messageId: 'noShadowGlobal',
          message: "'has' is already a global variable.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: `
declare const has: (environment: 'dev' | 'prod' | 'test') => boolean;
const fn = (has: string) => {};
      `,
      options: [
        {
          builtinGlobals: true,
        },
      ],
      languageOptions: {
        globals: {
          has: false,
        },
      },
      filename: 'decl.d.ts',
      errors: [
        {
          messageId: 'noShadow',
          message:
            "'has' is already declared in the upper scope on line 2 column 15.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
  ],
});
