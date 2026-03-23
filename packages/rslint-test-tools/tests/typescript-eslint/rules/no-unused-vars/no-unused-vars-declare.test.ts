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
  ],
});
