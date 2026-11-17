// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-mutable-exports', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'export const count = 1',
    },
    {
      name: 'run 0 valid 1',
      code: 'export function getCount() {}',
    },
    {
      name: 'run 0 valid 2',
      code: 'export class Counter {}',
    },
    {
      name: 'run 0 valid 3',
      code: 'export enum Direction { Up, Down }',
    },
    {
      name: 'run 0 valid 4',
      code: 'export interface Foo {}',
    },
    {
      name: 'run 0 valid 5',
      code: 'export type Foo = {}',
    },
    {
      name: 'run 0 valid 6',
      code: 'export const a = 1, b = 2',
    },
    {
      name: 'run 0 valid 7',
      code: 'export const { a, b } = obj',
    },
    {
      name: 'run 0 valid 8',
      code: 'export const [a, b] = arr',
    },
    {
      name: 'run 0 valid 9',
      code: 'export using x = getResource()',
    },
    {
      name: 'run 0 valid 10',
      code: 'export default function getCount() {}',
    },
    {
      name: 'run 0 valid 11',
      code: 'export default class Counter {}',
    },
    {
      name: 'run 0 valid 12',
      code: 'export default count = 1',
    },
    {
      name: 'run 0 valid 13',
      code: 'export default 42',
    },
    {
      name: 'run 0 valid 14',
      code: 'export default { x: 1 }',
    },
    {
      name: 'run 0 valid 15',
      code: 'export default [1, 2, 3]',
    },
    {
      name: 'run 0 valid 16',
      code: 'export default () => {}',
    },
    {
      name: 'run 0 valid 17',
      code: 'const count = 1\nexport { count }',
    },
    {
      name: 'run 0 valid 18',
      code: 'const count = 1\nexport { count as counter }',
    },
    {
      name: 'run 0 valid 19',
      code: 'const count = 1\nexport { count as default }',
    },
    {
      name: 'run 0 valid 20',
      code: 'const { a } = obj\nexport { a }',
    },
    {
      name: 'run 0 valid 21',
      code: 'const [a] = arr\nexport { a }',
    },
    {
      name: 'run 0 valid 22',
      code: 'const { a: b } = obj\nexport { b }',
    },
    {
      name: 'run 0 valid 23',
      code: 'const { a = 1 } = obj\nexport { a }',
    },
    {
      name: 'run 0 valid 24',
      code: 'const { ...rest } = obj\nexport { rest }',
    },
    {
      name: 'run 0 valid 25',
      code: 'const [, b] = arr\nexport { b }',
    },
    {
      name: 'run 0 valid 26',
      code: 'const { a: { b } } = obj\nexport { b }',
    },
    {
      name: 'run 0 valid 27',
      code: 'const count = 1\nexport default count',
    },
    {
      name: 'run 0 valid 28',
      code: 'const x = 1\nexport default (x)',
    },
    {
      name: 'run 0 valid 29',
      code: 'function getCount() {}\nexport { getCount }',
    },
    {
      name: 'run 0 valid 30',
      code: 'function getCount() {}\nexport { getCount as getCounter }',
    },
    {
      name: 'run 0 valid 31',
      code: 'function getCount() {}\nexport default getCount',
    },
    {
      name: 'run 0 valid 32',
      code: 'function getCount() {}\nexport { getCount as default }',
    },
    {
      name: 'run 0 valid 33',
      code: 'class Counter {}\nexport { Counter }',
    },
    {
      name: 'run 0 valid 34',
      code: 'class Counter {}\nexport { Counter as Count }',
    },
    {
      name: 'run 0 valid 35',
      code: 'class Counter {}\nexport default Counter',
    },
    {
      name: 'run 0 valid 36',
      code: 'class Counter {}\nexport { Counter as default }',
    },
    {
      name: 'run 0 valid 37',
      code: 'enum Direction { Up, Down }\nexport { Direction }',
    },
    {
      name: 'run 0 valid 38',
      code: 'type Foo = {}\nexport type { Foo }',
    },
    {
      name: 'run 0 valid 39',
      code: 'let x = 1\nexport { type x }',
    },
    {
      name: 'run 0 valid 40',
      code: "export { foo } from './foo'",
    },
    {
      name: 'run 0 valid 41',
      code: "export { foo as bar } from './foo'",
    },
    {
      name: 'run 0 valid 42',
      code: "export * from './foo'",
    },
    {
      name: 'run 0 valid 43',
      code: "export * as ns from './foo'",
    },
    {
      name: 'run 0 valid 44',
      code: "import { x } from './first'\nexport { x }",
    },
    {
      name: 'run 0 valid 45',
      code: "import { x } from './first'\nexport default x",
    },
    {
      name: 'run 0 valid 46',
      code: 'export { undeclared }',
    },
    {
      name: 'run 0 valid 47',
      code: 'export default undeclared',
    },
    {
      name: 'run 0 valid 48',
      code: 'export {}',
    },
    {
      name: 'run 0 valid 49',
      code: 'const x = 1\nexport = x',
    },
    {
      name: 'run 0 valid 50',
      code: 'const a = 1\nfunction b() {}\nexport { a, b }',
    },
    {
      name: 'run 0 valid 51',
      code: 'namespace Foo { export let x = 1 }',
    },
    {
      name: 'run 0 valid 52',
      code: 'module Foo { export let x = 1 }',
    },
    {
      name: 'run 0 valid 53',
      code: 'declare namespace Foo { export let x: number }',
    },
    {
      name: 'run 0 valid 54',
      code: "declare module 'foo' { export let x: number }",
    },
    {
      name: 'run 0 valid 55',
      code: 'namespace A { namespace B { export let x = 1 } }',
    },
    {
      name: 'run 0 valid 56',
      code: '{ let x = 1 }\nexport { x }',
    },
    {
      name: 'run 0 valid 57',
      code: '{ const x = 1 }\nexport { x }',
    },
    {
      name: 'run 0 valid 58',
      code: 'function foo() { var x = 1 }\nexport { x }',
    },
    {
      name: 'run 0 valid 59',
      code: 'const foo = () => { var x = 1 }\nexport { x }',
    },
    {
      name: 'run 0 valid 60',
      code: '(function() { var x = 1 })()\nexport { x }',
    },
    {
      name: 'run 0 valid 61',
      code: 'function Foo() {}\nnamespace Foo { export const bar = 1 }\nexport { Foo }',
    },
    {
      name: 'run 0 valid 62',
      code: 'const Foo = 1\nnamespace Foo {}\nexport { Foo }',
    },
    {
      name: 'run 0 valid 63',
      code: 'class Foo {}\nnamespace Foo {}\nexport { Foo }',
    },
    {
      name: 'run 0 valid 64',
      code: 'let x = 1\nexport default x + 1',
    },
    {
      name: 'run 0 valid 65',
      code: 'let x = 1\nexport default x!',
    },
    {
      name: 'run 0 valid 66',
      code: 'let x = 1\nexport default x as number',
    },
    {
      name: 'run 0 valid 67',
      code: 'var x = 1\nmodule.exports = x',
    },
    {
      name: 'run 0 valid 68',
      code: 'let x = 1\nmodule.exports = { x }',
    },
    {
      name: 'run 0 valid 69',
      code: 'const x = 1; export { x as "foo-bar" }',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'export let count = 1',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'export var count = 1',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'export let a = 1, b = 2',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'export var a = 1, b = 2',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'export let { a, b } = obj',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'export var { a, b } = obj',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'export let [a, b] = arr',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'export var [a, b] = arr',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: "export let x: string = 'hello'",
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'export declare let x: number',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'let count = 1\nexport { count }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'var count = 1\nexport { count }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: 'let count = 1\nexport { count as counter }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'var count = 1\nexport { count as counter }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'let count = 1\nexport { count as default }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: 'let count = 1\nexport default count',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: 'var count = 1\nexport default count',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: 'let x\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: 'var x\nexport default x',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      name: 'run 0 invalid 19',
      code: 'let { a } = obj\nexport { a }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 20',
      code: 'let [a] = arr\nexport { a }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 21',
      code: 'var { a: { b } } = obj\nexport { b }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 22',
      code: 'var { x } = obj\nexport default x',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 23',
      code: 'let { ...rest } = obj\nexport { rest }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 24',
      code: 'let { a: b } = obj\nexport { b }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 0 invalid 25',
      code: 'var { a = 1 } = obj\nexport { a }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'run 0 invalid 26',
      code: 'let [, b] = arr\nexport { b }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 27',
      code: 'var [a, ...rest] = arr\nexport { rest }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 28',
      code: 'let { a: { b: [c] } } = obj\nexport { c }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 29',
      code: 'let a = 1\nlet b = 2\nexport { a, b }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 0 invalid 30',
      code: 'let a = 1\nconst b = 2\nexport { a, b }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 0 invalid 31',
      code: 'export let x = 1\nexport { x as y }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 17,
        },
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      name: 'run 0 invalid 32',
      code: 'let x = 1\nexport { x, x as y, x as z }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 0 invalid 33',
      code: 'export { x }\nvar x = 1',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 0 invalid 34',
      code: 'let x = 1; export { x as "foo-bar" }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 0 invalid 35',
      code: 'let x = 1 as const\nexport default x',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 0 invalid 36',
      code: '{ var x = 1 }\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 37',
      code: 'if (true) { var x = 1 }\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 38',
      code: 'for (var x = 0; x < 1; x++) {}\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 39',
      code: 'for (var x of []) {}\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 40',
      code: 'for (var x in {}) {}\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 41',
      code: 'try { var x = 1 } catch {}\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 42',
      code: '{ { { var x = 1 } } }\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 43',
      code: '{ var x = 1 }\nexport default x',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 44',
      code: 'switch (1) { case 1: var x = 1; break; }\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 45',
      code: 'while (false) { var x = 1 }\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 46',
      code: 'do { var x = 1 } while (false)\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 47',
      code: 'var Foo = 1\nnamespace Foo {}\nexport { Foo }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 48',
      code: 'declare var x: number\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'var' binding, use 'const' instead.",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 49',
      code: 'declare let x: number\nexport { x }',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 50',
      code: 'let x = 1\nexport default (x)',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 0 invalid 51',
      code: 'let x = 1\nexport default (((x)))',
      errors: [
        {
          messageId: 'noMutableExports',
          message: "Exporting mutable 'let' binding, use 'const' instead.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
  ],
});
