import { test } from '../utils.js';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();

ruleTester.run('no-mutable-exports', null as never, {
  valid: [
    // === Direct exports ===
    test({ code: 'export const count = 1' }),
    test({ code: 'export function getCount() {}' }),
    test({ code: 'export class Counter {}' }),
    test({ code: 'export enum Direction { Up, Down }' }),
    test({ code: 'export interface Foo {}' }),
    test({ code: 'export type Foo = {}' }),
    test({ code: 'export const a = 1, b = 2' }),
    test({ code: 'export const { a, b } = obj' }),
    test({ code: 'export const [a, b] = arr' }),
    test({ code: 'export using x = getResource()' }),

    // === Default exports ===
    test({ code: 'export default function getCount() {}' }),
    test({ code: 'export default class Counter {}' }),
    test({ code: 'export default count = 1' }),
    test({ code: 'export default 42' }),
    test({ code: 'export default { x: 1 }' }),
    test({ code: 'export default [1, 2, 3]' }),
    test({ code: 'export default () => {}' }),

    // === Named exports with const ===
    test({ code: 'const count = 1\nexport { count }' }),
    test({ code: 'const count = 1\nexport { count as counter }' }),
    test({ code: 'const count = 1\nexport { count as default }' }),
    test({ code: 'const { a } = obj\nexport { a }' }),
    test({ code: 'const [a] = arr\nexport { a }' }),
    test({ code: 'const { a: b } = obj\nexport { b }' }),
    test({ code: 'const { a = 1 } = obj\nexport { a }' }),
    test({ code: 'const { ...rest } = obj\nexport { rest }' }),
    test({ code: 'const [, b] = arr\nexport { b }' }),
    test({ code: 'const { a: { b } } = obj\nexport { b }' }),

    // === Default export of const ===
    test({ code: 'const count = 1\nexport default count' }),
    // Parenthesized const (parens stripped)
    test({ code: 'const x = 1\nexport default (x)' }),

    // === Function/class then export ===
    test({ code: 'function getCount() {}\nexport { getCount }' }),
    test({
      code: 'function getCount() {}\nexport { getCount as getCounter }',
    }),
    test({ code: 'function getCount() {}\nexport default getCount' }),
    test({ code: 'function getCount() {}\nexport { getCount as default }' }),
    test({ code: 'class Counter {}\nexport { Counter }' }),
    test({ code: 'class Counter {}\nexport { Counter as Count }' }),
    test({ code: 'class Counter {}\nexport default Counter' }),
    test({ code: 'class Counter {}\nexport { Counter as default }' }),

    // === Enum then export ===
    test({ code: 'enum Direction { Up, Down }\nexport { Direction }' }),

    // === Type-only exports ===
    test({ code: 'type Foo = {}\nexport type { Foo }' }),
    test({ code: 'let x = 1\nexport { type x }' }),

    // === Re-exports ===
    test({ code: "export { foo } from './foo'" }),
    test({ code: "export { foo as bar } from './foo'" }),
    test({ code: "export * from './foo'" }),
    test({ code: "export * as ns from './foo'" }),

    // === Import then re-export ===
    test({ code: "import { x } from './first'\nexport { x }" }),
    test({ code: "import { x } from './first'\nexport default x" }),

    // === Undeclared / empty ===
    test({ code: 'export { undeclared }' }),
    test({ code: 'export default undeclared' }),
    test({ code: 'export {}' }),

    // === TypeScript: export = ===
    test({ code: 'const x = 1\nexport = x' }),

    // === Mixed specifiers: all const/func ===
    test({ code: 'const a = 1\nfunction b() {}\nexport { a, b }' }),

    // === Namespace/module: export let inside is NOT an ES module export ===
    test({ code: 'namespace Foo { export let x = 1 }' }),
    test({ code: 'module Foo { export let x = 1 }' }),
    test({ code: 'declare namespace Foo { export let x: number }' }),
    test({ code: "declare module 'foo' { export let x: number }" }),
    test({ code: 'namespace A { namespace B { export let x = 1 } }' }),

    // === Hoisting boundaries ===
    test({ code: '{ let x = 1 }\nexport { x }' }),
    test({ code: '{ const x = 1 }\nexport { x }' }),
    test({ code: 'function foo() { var x = 1 }\nexport { x }' }),
    test({ code: 'const foo = () => { var x = 1 }\nexport { x }' }),
    test({ code: '(function() { var x = 1 })()\nexport { x }' }),

    // === Declaration merging: all immutable ===
    test({
      code: 'function Foo() {}\nnamespace Foo { export const bar = 1 }\nexport { Foo }',
    }),
    test({ code: 'const Foo = 1\nnamespace Foo {}\nexport { Foo }' }),
    test({ code: 'class Foo {}\nnamespace Foo {}\nexport { Foo }' }),

    // === Non-identifier default export expressions ===
    test({ code: 'let x = 1\nexport default x + 1' }),
    test({ code: 'let x = 1\nexport default x!' }),
    test({ code: 'let x = 1\nexport default x as number' }),

    // === CommonJS (not handled by this rule) ===
    test({ code: 'var x = 1\nmodule.exports = x' }),
    test({ code: 'let x = 1\nmodule.exports = { x }' }),

    // === String literal export name with const ===
    test({ code: 'const x = 1; export { x as "foo-bar" }' }),
  ],
  invalid: [
    // === Direct export let/var ===
    test({
      code: 'export let count = 1',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'export var count = 1',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'export let a = 1, b = 2',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'export var a = 1, b = 2',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    // Direct export with destructuring
    test({
      code: 'export let { a, b } = obj',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'export var { a, b } = obj',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'export let [a, b] = arr',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'export var [a, b] = arr',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    // TypeScript: let with type annotation
    test({
      code: "export let x: string = 'hello'",
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    // TypeScript: declare let
    test({
      code: 'export declare let x: number',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),

    // === Named exports referencing let/var ===
    test({
      code: 'let count = 1\nexport { count }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'var count = 1\nexport { count }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let count = 1\nexport { count as counter }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'var count = 1\nexport { count as counter }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let count = 1\nexport { count as default }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),

    // === Default export of let/var ===
    test({
      code: 'let count = 1\nexport default count',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'var count = 1\nexport default count',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    // Uninitialized
    test({
      code: 'let x\nexport { x }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'var x\nexport default x',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),

    // === Destructuring then named/default export ===
    test({
      code: 'let { a } = obj\nexport { a }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let [a] = arr\nexport { a }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'var { a: { b } } = obj\nexport { b }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'var { x } = obj\nexport default x',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let { ...rest } = obj\nexport { rest }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let { a: b } = obj\nexport { b }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'var { a = 1 } = obj\nexport { a }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let [, b] = arr\nexport { b }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'var [a, ...rest] = arr\nexport { rest }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let { a: { b: [c] } } = obj\nexport { c }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),

    // === Multiple errors ===
    test({
      code: 'let a = 1\nlet b = 2\nexport { a, b }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let a = 1\nconst b = 2\nexport { a, b }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'export let x = 1\nexport { x as y }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let x = 1\nexport { x, x as y, x as z }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),

    // === Declaration after export (var hoists, we still find it) ===
    test({
      code: 'export { x }\nvar x = 1',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),

    // === ES2022: string literal export name with let ===
    test({
      code: 'let x = 1; export { x as "foo-bar" }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    // Type assertion doesn't change mutability (still let)
    test({
      code: 'let x = 1 as const\nexport default x',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),

    // Hoisted var (var inside block hoists to module scope)
    test({
      code: '{ var x = 1 }\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'if (true) { var x = 1 }\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'for (var x = 0; x < 1; x++) {}\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'for (var x of []) {}\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'for (var x in {}) {}\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'try { var x = 1 } catch {}\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: '{ { { var x = 1 } } }\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: '{ var x = 1 }\nexport default x',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    // switch / while / do-while / label / else / catch / finally
    test({
      code: 'switch (1) { case 1: var x = 1; break; }\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'while (false) { var x = 1 }\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'do { var x = 1 } while (false)\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    // Declaration merging: mutable binding
    test({
      code: 'var Foo = 1\nnamespace Foo {}\nexport { Foo }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    // declare var/let
    test({
      code: 'declare var x: number\nexport { x }',
      errors: [
        { message: "Exporting mutable 'var' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'declare let x: number\nexport { x }',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),

    // Parenthesized export default
    test({
      code: 'let x = 1\nexport default (x)',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
    test({
      code: 'let x = 1\nexport default (((x)))',
      errors: [
        { message: "Exporting mutable 'let' binding, use 'const' instead." },
      ],
    }),
  ],
});
