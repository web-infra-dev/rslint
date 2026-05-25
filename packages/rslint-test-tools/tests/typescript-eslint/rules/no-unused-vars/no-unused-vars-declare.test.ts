import { RuleTester } from '@typescript-eslint/rule-tester';
import { getFixturesRootDir } from '../../RuleTester';

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: getFixturesRootDir(),
    },
  },
});

ruleTester.run('no-unused-vars', {
  valid: [
    // --- basic usage ---
    `const foo = 5; console.log(foo);`,
    // shorthand property counts as usage
    `function test(stats: string) { console.log({ stats }); } test("ok");`,
    `function foo() {} foo();`,
    `function foo(bar: number) { console.log(bar); } foo(1);`,
    `try {} catch (e) { console.log(e); }`,
    { code: `export const foo = 1;` },
    // type-annotated variable that IS used
    `const bar: number = 1; console.log(bar);`,

    // --- options (array format) ---
    {
      code: `const { foo, ...rest } = { foo: 1, bar: 2 }; console.log(rest);`,
      options: [{ ignoreRestSiblings: true }],
    },
    {
      code: `const _foo = 1;`,
      options: [{ varsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(_bar: number) {}
foo(1);
      `,
      options: [{ argsIgnorePattern: '^_' }],
    },
    {
      code: `
function foo(bar: number) {}
foo(1);
      `,
      options: [{ args: 'none' }],
    },
    {
      code: `try {} catch (e) {}`,
      options: [{ caughtErrors: 'none' }],
    },
    {
      code: `try {} catch (_err) {}`,
      options: [{ caughtErrorsIgnorePattern: '^_' }],
    },

    // --- issue #555: declare function overloads ---
    {
      code: `
export type NormalizedConfig = { key: string };
declare function getNormalizedConfig(): NormalizedConfig;
declare function getNormalizedConfig(options: { environment: string }): NormalizedConfig;
export { getNormalizedConfig };
      `,
    },
    {
      code: `
declare function doSomething(options: { a: string }): void;
export { doSomething };
      `,
    },
    {
      code: `
declare function getNormalizedConfig(): string;
declare function getNormalizedConfig(options: { env: string }): string;
getNormalizedConfig();
      `,
    },
    `declare function foo(): void; foo();`,
    {
      code: `
export function foo(a: number): number;
export function foo(a: string): string;
export function foo(a: number | string): number | string {
  return a;
}
      `,
    },
    {
      code: `
export function foo(): void;
export function foo(): void;
export function foo(): void {}
      `,
    },
    {
      code: `
declare function process(input: string, options: { verbose: boolean }): void;
export { process };
      `,
    },
    {
      code: `
declare function withRest(...args: any[]): void;
export { withRest };
      `,
    },
    {
      code: `
function foo(): void;
function foo(): void {}
foo();
      `,
    },

    // --- declare namespace ---
    {
      code: `
declare namespace MyNS {
  function nsFunc(param: string): void;
  var nsVar: string;
}
console.log(MyNS);
      `,
    },
    {
      code: `
declare module 'some-module' {
  function moduleFunc(arg: string): void;
}
      `,
    },

    // --- bodyless declaration params ---
    {
      code: `
abstract class AbstractBase {
  abstract doSomething(input: string, options: { verbose: boolean }): void;
}
export { AbstractBase };
      `,
    },
    {
      code: `
class MyClass {
  method(a: number): number;
  method(a: string): string;
  method(a: number | string): number | string {
    return a;
  }
}
export { MyClass };
      `,
    },
    {
      code: `
export interface IProcessor {
  process(input: string, options: { debug: boolean }): void;
}
      `,
    },
    // function type literal params (type-level, never reported)
    {
      code: `
export interface Hot {
  on: <Data = any>(event: string, cb: (data: Data) => void) => void;
}
      `,
    },
    // call signature params
    {
      code: `
export interface Callable {
  (x: number, y: string): boolean;
}
      `,
    },
    // construct signature params
    {
      code: `
export interface Constructable {
  new (name: string): object;
}
      `,
    },
    // function type in type alias
    {
      code: `export type Handler = (event: string, data: unknown) => void;`,
    },
    // index signature param
    {
      code: `export interface Dict { [key: string]: unknown; }`,
    },
    // declare global (global scope augmentation, never reported)
    {
      code: `declare global { const BUILD_HASH: string; }`,
    },

    // --- class/interface/type/enum: used ---
    `class Foo {} new Foo();`,
    { code: `export class Foo {}` },
    { code: `export interface Bar { x: number; }` },
    { code: `export type Str = string;` },
    `enum Color { Red, Blue } console.log(Color.Red);`,
    { code: `export enum Color { Red, Blue }` },

    {
      code: `
export class MyClass {
  constructor(a: number);
  constructor(a: string);
  constructor(a: number | string) { console.log(a); }
}
      `,
    },

    // --- export declare / generic ---
    {
      code: `export declare function exportDeclare(x: number): void;`,
    },
    {
      code: `
declare function genericFunc<T>(input: T): T;
export { genericFunc };
      `,
    },

    // --- export class/function with used params ---
    {
      code: `
export class Baz {
  value: number;
  constructor(a: number) { this.value = a; }
}
      `,
    },
    { code: `export function bar(x: number) { return x; }` },
    { code: `export namespace ExportedNS { export const x = 1; }` },

    // --- scope ---
    {
      code: `
const x = 1;
function foo(x: number) { return x; }
console.log(x);
foo(2);
      `,
    },

    // --- destructuring: used ---
    `const { a } = { a: 1 }; console.log(a);`,
    `const [p] = [1]; console.log(p);`,
    // --- namespace import: used ---
    `import * as path from "path"; console.log(path.join("a", "b"));`,
    // import equals: used
    `import path = require("path"); console.log(path.join("a", "b"));`,
    // parameter destructuring: used
    `function foo({ a }: { a: number }) { console.log(a); } foo({ a: 1 });`,
    // parameter destructuring + argsIgnorePattern
    {
      code: `function foo({ _a, b }: { _a: number; b: number }) { console.log(b); } foo({ _a: 1, b: 2 });`,
      options: [{ argsIgnorePattern: '^_' }],
    },
    // parameter destructuring + args "none"
    {
      code: `function foo({ a }: { a: number }) {} foo({ a: 1 });`,
      options: [{ args: 'none' }],
    },

    // --- destructuredArrayIgnorePattern ---
    {
      code: `const [_a, b] = [1, 2]; console.log(b);`,
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    {
      code: `const [_a, _b] = [1, 2];`,
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    // varsIgnorePattern alone also saves array destructured
    {
      code: `const [_a] = [1];`,
      options: [{ varsIgnorePattern: '^_' }],
    },
    // nested array destructuring
    {
      code: `const [[_a], b] = [[1], 2]; console.log(b);`,
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },
    // parameter array destructuring
    {
      code: `function foo([_a, b]: number[]) { console.log(b); } foo([1, 2]);`,
      options: [{ destructuredArrayIgnorePattern: '^_' }],
    },

    // --- ignoreClassWithStaticInitBlock ---
    {
      code: `class Foo { static { console.log("init"); } }`,
      options: [{ ignoreClassWithStaticInitBlock: true }],
    },
    // varsIgnorePattern can independently save a class with static block
    {
      code: `class Foo { static {} }`,
      options: [{ ignoreClassWithStaticInitBlock: false, varsIgnorePattern: '^Foo' }],
    },

    // --- ignoreUsingDeclarations ---
    {
      code: `using resource = {} as any;`,
      options: [{ ignoreUsingDeclarations: true }],
    },
    {
      code: `await using resource = {} as any;`,
      options: [{ ignoreUsingDeclarations: true }],
    },

    // --- write-only destructuring assignment: used after reassignment ---
    `let a: any, b: any; [a, b] = [1, 2]; console.log(a, b);`,
    `let a: any; ({ a } = { a: 1 }); console.log(a);`,
    `let rest: any; [, ...rest] = [1, 2, 3]; console.log(rest);`,
    `let rest: any; ({ ...rest } = { a: 1, b: 2 } as any); console.log(rest);`,
    `let a: any; [{ a }] = [{ a: 1 }] as any; console.log(a);`,
    `let b: any; ({ a: b } = { a: 1 } as any); console.log(b);`,
    `let a: any; [(a)] = [1] as any; console.log(a);`,
    // for-of/for-in: variable used in body
    `let x: any; for (x of [1, 2]) { console.log(x); }`,
    `let k: any; for (k in { a: 1 }) { console.log(k); }`,
    `let a: any; for ({ a } of [{ a: 1 }] as any) { console.log(a); }`,
    // object spread assignment: used
    `let rest: any; ({ ...rest } = { a: 1, b: 2 } as any); console.log(rest);`,
    // nested mixed: used
    `let a: any; [{ a }] = [{ a: 1 }] as any; console.log(a);`,
    `let b: any; ({ x: [b] } = { x: [1] } as any); console.log(b);`,
    // renamed property + default value: used
    `let b: any; ({ a: b = 5 } = {} as any); console.log(b);`,
    // deeply nested: used
    `let a: any; [[[[a]]]] = [[[[1]]]] as any; console.log(a);`,
    `let a: any; [{ x: [{ a }] }] = [{ x: [{ a: 1 }] }] as any; console.log(a);`,
    // default value in destructuring assignment: used
    `let a: any; [a = 5] = [] as any; console.log(a);`,
    // computed property name: key IS a read
    `const key = "x"; let val: any; ({ [key]: val } = { x: 1 } as any); console.log(val);`,
    // property/element access assignment: object IS a read
    `const obj = { b: 0 }; obj.b = 1; console.log(obj);`,
    `const arr = [0]; arr[0] = 1; console.log(arr);`,
    // side-effect import
    `import "path";`,
    // import re-export
    `import { join } from "path"; export { join };`,
    `import { join } from "path"; export { join as myJoin };`,
    // namespace re-export
    `import * as path from "path"; export { path };`,
    // export default import
    `import { join } from "path"; export default join;`,
    // direct re-export (no local binding)
    `export { join } from "path";`,
    // local function re-export
    `function foo() {} export { foo };`,
    // self-assignment: result consumed
    `let a = 0; a = a + 1; console.log(a);`,
    `let a = 0; a++; console.log(a);`,
    // namespace used externally
    `namespace Foo { export const Bar = 1; } console.log(Foo.Bar);`,
    // setter param: required by syntax
    {
      code: `
export const obj = {
  set foo(a: number) {}
};
      `,
    },
  ],
  invalid: [
    // --- basic unused ---
    {
      code: `const foo = 5;`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `function foo() {}`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // unused function with param → report both function name AND param
    {
      code: `function foo(bar: number) {}`,
      errors: [{ messageId: 'unusedVar' }, { messageId: 'unusedVar' }],
    },
    // unused function with multiple params → report all
    {
      code: `function unused(a: number, b: string) {}`,
      errors: [
        { messageId: 'unusedVar' },
        { messageId: 'unusedVar' },
        { messageId: 'unusedVar' },
      ],
    },
    {
      code: `try {} catch (e) {}`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // type-annotated but unused variable
    {
      code: `const bar: number = 1;`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `let foo = 5; foo = 10;`,
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- usedOnlyAsType ---
    {
      code: `const foo = 1; type Bar = typeof foo; export type { Bar };`,
      options: [{ vars: 'all' }],
      errors: [{ messageId: 'usedOnlyAsType' }],
    },

    // --- options (array format) ---
    {
      code: `const foo = 1;`,
      options: [{ varsIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `
const _foo = 1; console.log(_foo);
      `,
      options: [{ varsIgnorePattern: '^_', reportUsedIgnorePattern: true }],
      errors: [{ messageId: 'usedIgnoredVar' }],
    },
    // reportUsedIgnorePattern applies to argsIgnorePattern
    {
      code: `
function foo(_x: number) { return _x; }
foo(1);
      `,
      options: [{ argsIgnorePattern: '^_', reportUsedIgnorePattern: true }],
      errors: [{ messageId: 'usedIgnoredVar' }],
    },
    // reportUsedIgnorePattern applies to caughtErrorsIgnorePattern
    {
      code: `try { throw 1; } catch (_e) { console.log(_e); }`,
      options: [{ caughtErrorsIgnorePattern: '^_', reportUsedIgnorePattern: true }],
      errors: [{ messageId: 'usedIgnoredVar' }],
    },
    // varsIgnorePattern must NOT apply to params
    {
      code: `
function foo(_x: number) {}
foo(1);
      `,
      options: [{ varsIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },
    // varsIgnorePattern must NOT apply to catch
    {
      code: `try {} catch (_e) {}`,
      options: [{ varsIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `export function foo(bar: number) {}`,
      options: [{ argsIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `try {} catch (err) {}`,
      options: [{ caughtErrorsIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- declare function ---
    {
      code: `declare function unusedFunc(): void;`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `
declare function unusedOverload(): void;
declare function unusedOverload(x: number): void;
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `
declare function typedFunc(): void;
type FuncType = typeof typedFunc;
export type { FuncType };
      `,
      errors: [{ messageId: 'usedOnlyAsType' }],
    },

    // --- unused declare namespace ---
    {
      code: `
declare namespace UnusedNS {
  function inner(): void;
}
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // empty declare namespace
    {
      code: `declare namespace Rspack {}`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // empty namespace (non-declare)
    {
      code: `namespace Rspack2 {}`,
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- unused class/interface/type/enum ---
    {
      code: `class UnusedClass {}`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `interface UnusedInterface { x: number; }`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `type UnusedType = string;`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `enum UnusedEnum { A, B }`,
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- isExported fix ---
    {
      code: `
export class UnusedCtorParam {
  constructor(a: number) {}
}
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `export function unusedFnParam(x: number) {}`,
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- scope ---
    {
      code: `
declare function other(x: number): void;
export { other };
export function scopeTest(x: number) {}
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // --- after-used: only report params after last used (default) ---
    {
      code: `
export function foo(used: number, unused1: string, unused2: boolean) {
  return used;
}
      `,
      errors: [{ messageId: 'unusedVar' }, { messageId: 'unusedVar' }],
    },
    // after-used: middle param used, only report after it
    {
      code: `
export function qux(a: number, b: string, c: boolean) {
  console.log(b);
}
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // after-used: all params unused — report all
    {
      code: `export function bar(a: number, b: string, c: boolean) {}`,
      errors: [
        { messageId: 'unusedVar' },
        { messageId: 'unusedVar' },
        { messageId: 'unusedVar' },
      ],
    },
    // args: "all" — report ALL unused params regardless of position
    {
      code: `
export function qux(a: number, b: string, c: boolean) {
  console.log(b);
}
      `,
      options: [{ args: 'all' }],
      errors: [{ messageId: 'unusedVar' }, { messageId: 'unusedVar' }],
    },
    // after-used: default value param acts as boundary
    {
      code: `export const fn = (_a: string, _b: number, _c = {}) => {};`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // --- destructuring: unused element ---
    {
      code: `const { a, b } = { a: 1, b: 2 }; console.log(a);`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `const [p, q] = [1, 2]; console.log(p);`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // namespace import: unused
    {
      code: `import * as ns from "./foo";`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // import equals: unused
    {
      code: `import path = require("path");`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // parameter destructuring: unused element
    {
      code: `
function foo({ a, b }: { a: number; b: string }) { console.log(a); }
foo({ a: 1, b: "x" });
      `,
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- destructuredArrayIgnorePattern ---
    // pattern does not match
    {
      code: `const [foo] = [1];`,
      options: [{ destructuredArrayIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },
    // object destructuring NOT affected
    {
      code: `const { _a } = { _a: 1 };`,
      options: [{ destructuredArrayIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },
    // object inside array NOT affected
    {
      code: `
const array = [{}];
const [{ _a, foo }] = array;
console.log(foo);
      `,
      options: [{ destructuredArrayIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- ignoreClassWithStaticInitBlock ---
    // class WITHOUT static block → still reported
    {
      code: `class Foo { static x = 1; }`,
      options: [{ ignoreClassWithStaticInitBlock: true }],
      errors: [{ messageId: 'unusedVar' }],
    },
    // class WITH static block → reported when option is default/false
    {
      code: `class Foo { static { console.log("init"); } }`,
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- ignoreUsingDeclarations ---
    // using → reported when option is default
    {
      code: `using resource = {} as any;`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // const NOT affected by ignoreUsingDeclarations
    {
      code: `const foo = 1;`,
      options: [{ ignoreUsingDeclarations: true }],
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- write-only destructuring assignment ---
    {
      code: `
let a: any, b: any;
[a, b] = [1, 2];
      `,
      errors: [{ messageId: 'unusedVar' }, { messageId: 'unusedVar' }],
    },
    {
      code: `
let a: any, b: any;
({ a, b } = { a: 1, b: 2 });
      `,
      errors: [{ messageId: 'unusedVar' }, { messageId: 'unusedVar' }],
    },
    // nested write-only
    {
      code: `
let a: any;
[[a]] = [[1]] as any;
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // parenthesized assignment target: write-only
    {
      code: `
let a: any;
[(a)] = [1] as any;
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // for-of write-only
    {
      code: `
let x: any;
for (x of [1, 2]) {}
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // for-in write-only
    {
      code: `
let k: any;
for (k in { a: 1 }) {}
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // for-of object destructuring: write-only
    {
      code: `
let a: any;
for ({ a } of [{ a: 1 }] as any) {}
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // object spread assignment: write-only
    {
      code: `
let rest: any;
({ ...rest } = { a: 1 } as any);
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // renamed property: write-only
    {
      code: `
let b: any;
({ a: b } = { a: 1 } as any);
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // nested object in array: write-only
    {
      code: `
let a: any;
[{ a }] = [{ a: 1 }] as any;
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // nested array in object: write-only
    {
      code: `
let b: any;
({ x: [b] } = { x: [1] } as any);
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // deeply nested: write-only
    {
      code: `
let a: any;
[[[[a]]]] = [[[[1]]]] as any;
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // default value in destructuring assignment: write-only
    {
      code: `
let a: any;
[a = 5] = [] as any;
      `,
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- pattern isolation ---
    // argsIgnorePattern should NOT apply to vars
    {
      code: `const _x = 1;`,
      options: [{ argsIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },
    // caughtErrorsIgnorePattern should NOT apply to params
    {
      code: `export function foo(_x: number) {}`,
      options: [{ caughtErrorsIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },
    // destructuredArrayIgnorePattern should NOT apply to plain params
    {
      code: `export function foo(_x: number) {}`,
      options: [{ destructuredArrayIgnorePattern: '^_' }],
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- ignoreClassWithStaticInitBlock ---
    // class with static method (not block) → still reported
    {
      code: `
class Foo {
  static bar() {}
}
      `,
      options: [{ ignoreClassWithStaticInitBlock: true }],
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- ignoreUsingDeclarations ---
    // let NOT affected
    {
      code: `let foo = 1;`,
      options: [{ ignoreUsingDeclarations: true }],
      errors: [{ messageId: 'unusedVar' }],
    },
    // await using default → reported
    {
      code: `await using resource = {} as any;`,
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- import fix: partial removal ---
    {
      code: `import { join, resolve } from "path"; console.log(join("a", "b"));`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // partial re-export: resolve unused
    {
      code: `import { join, resolve } from "path"; export { join };`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // default import unused
    {
      code: `import Foo from "./foo";`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // type import unused
    {
      code: `import type { Foo } from "./foo";`,
      errors: [{ messageId: 'unusedVar' }],
    },

    // --- self-assignment: variable only modifies itself ---
    {
      code: `var a = 0; a = a + 1;`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `var a = 0; a++;`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `var a = 0; a += 1;`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `function foo(a: number) { a = a + 1; } foo(1);`,
      errors: [{ messageId: 'unusedVar' }],
    },
    {
      code: `var a = 3; a = a * 5 + 6;`,
      errors: [{ messageId: 'unusedVar' }],
    },
    // --- namespace self-reference ---
    {
      code: `
namespace Foo {
  export const Bar = 1;
  console.log(Foo.Bar);
}
      `,
      errors: [{ messageId: 'unusedVar' }],
    },
    // --- recursive function ---
    {
      code: `function foox() { return foox(); }`,
      errors: [{ messageId: 'unusedVar' }],
    },
  ],
});
