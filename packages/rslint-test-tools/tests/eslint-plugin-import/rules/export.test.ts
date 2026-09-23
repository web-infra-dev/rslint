import { RuleTester } from '../rule-tester.js';
import { testFixturePath } from '../utils.js';

// Upstream: eslint-plugin-import v2.32.0 tests/src/rules/export.js and docs/rules/export.md.
// Parser/version variants are deduplicated; SYNTAX_CASES are expanded.
new RuleTester().run('export', null as never, {
  valid: [
    // JavaScript and SYNTAX_CASES
    // Skipped: the project rejects this malformed dependency before rules run.
    // import "./malformed.js"
    {
      code: 'var foo = "foo"; export default foo;',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export var foo = "foo"; export var bar = "bar";',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export var foo = "foo", bar = "bar";',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export var { foo, bar } = object;',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export var [ foo, bar ] = array;',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'let foo; export { foo, foo as bar }',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'let bar; export { bar }; export * from "./export-all"',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export * from "./export-all"',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export * from "./does-not-exist"',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export default foo; export * from "./bar"',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'for (let { foo, bar } of baz) {}',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'for (let [ foo, bar ] of baz) {}',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'const { x, y } = bar',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'const { x, y, ...z } = bar',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'let x; export { x }',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'let x; export { x as y }',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export const x = null',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export var x = null',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export let x = null',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export default x',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export default class x {}',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'import json from "./data.json"',
      filename: testFixturePath('export-rule/consumer.ts'),
      settings: { 'import/extensions': ['.js'] },
    },
    {
      code: 'import foo from "./foobar.json";',
      filename: testFixturePath('export-rule/consumer.ts'),
      settings: { 'import/extensions': ['.js'] },
    },
    {
      code: 'import foo from "./foobar";',
      filename: testFixturePath('export-rule/consumer.ts'),
      settings: { 'import/extensions': ['.js'] },
    },
    {
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      filename: testFixturePath('export-rule/consumer.ts'),
      settings: { 'import/ignore': ['foo'] },
    },
    {
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      filename: testFixturePath('export-rule/consumer.ts'),
      settings: { 'import/ignore': ['foo'] },
    },
    {
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'import { foo } from "./ignore.invalid.extension"',
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
        import * as A from './named-export-collision/a';
        import * as B from './named-export-collision/b';

        export { A, B };
      `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
        export * as A from './named-export-collision/a';
        export * as B from './named-export-collision/b';
      `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
        export default function foo(param: string): boolean;
        export default function foo(param: string, param1: number): boolean;
        export default function foo(param: string, param1?: number): boolean {
          return param && param1;
        }
      `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
        export default function foo(param: string): boolean;
        export default function foo(param: string, param1?: number): boolean {
          return param && param1;
        }
      `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    // TypeScript
    {
      code: `
            export const Foo = 1;
            export type Foo = number;
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            export const Foo = 1;
            export interface Foo {}
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            export function fff(a: string);
            export function fff(a: number);
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            export function fff(a: string);
            export function fff(a: number);
            export function fff(a: string|number) {};
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            export const Bar = 1;
            export namespace Foo {
              export const Bar = 1;
            }
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            export type Bar = string;
            export namespace Foo {
              export type Bar = string;
            }
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            export const Bar = 1;
            export type Bar = string;
            export namespace Foo {
              export const Bar = 1;
              export type Bar = string;
            }
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            export namespace Foo {
              export const Foo = 1;
              export namespace Bar {
                export const Foo = 2;
              }
              export namespace Baz {
                export const Foo = 3;
              }
            }
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
              export class Foo { }
              export namespace Foo { }
              export namespace Foo {
                export class Bar {}
              }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
              export function Foo();
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
              export function Foo(a: string);
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
              export function Foo(a: string);
              export function Foo(a: number);
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
              export enum Foo { }
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: 'export * from "./file1.ts"',
      filename: testFixturePath('export-rule/typescript-d-ts/file-2.ts'),
    },
    {
      code: `
              export * as A from './named-export-collision/a';
              export * as B from './named-export-collision/b';
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            declare module "a" {
              const Foo = 1;
              export {Foo as default};
            }
            declare module "b" {
              const Bar = 2;
              export {Bar as default};
            }
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            declare module "a" {
              const Foo = 1;
              export {Foo as default};
            }
            const Bar = 2;
            export {Bar as default};
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
    },
    {
      code: `
            export * from './module';
          `,
      filename: testFixturePath('export-rule/export-star-4/index.ts'),
      settings: { 'import/extensions': ['.js', '.ts', '.jsx'] },
    },
    // Documentation
  ],
  invalid: [
    // JavaScript and SYNTAX_CASES
    {
      code: 'let foo; export { foo }; export * from "./export-all"',
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'foo'.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 22,
        },
        {
          message: "Multiple exports of name 'foo'.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    // Skipped: tsgo accepts the top-level return that Espree rejects.
    // export * from "./malformed.js"
    // Parse errors in imported module './malformed.js': 'return' outside of function (1:1)
    {
      code: 'export * from "./default-export"',
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "No named exports found in module './default-export'.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'let foo; export { foo as "foo" }; export * from "./export-all"',
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'foo'.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 31,
        },
        {
          message: "Multiple exports of name 'foo'.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      code: `
        export default function a(): void;
        export default function a() {}
        export { x as default };
      `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: 'Multiple default exports.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 39,
        },
        {
          message: 'Multiple default exports.',
          line: 4,
          column: 23,
          endLine: 4,
          endColumn: 30,
        },
      ],
    },
    // TypeScript
    {
      code: `
            export type Foo = string;
            export type Foo = number;
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'Foo'.",
          line: 2,
          column: 25,
          endLine: 2,
          endColumn: 28,
        },
        {
          message: "Multiple exports of name 'Foo'.",
          line: 3,
          column: 25,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    {
      code: `
            export const a = 1
            export namespace Foo {
              export const a = 2;
              export const a = 3;
            }
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'a'.",
          line: 4,
          column: 28,
          endLine: 4,
          endColumn: 29,
        },
        {
          message: "Multiple exports of name 'a'.",
          line: 5,
          column: 28,
          endLine: 5,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
            declare module 'foo' {
              const Foo = 1;
              export default Foo;
              export default Foo;
            }
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: 'Multiple default exports.',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 34,
        },
        {
          message: 'Multiple default exports.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 34,
        },
      ],
    },
    {
      code: `
            export namespace Foo {
              export namespace Bar {
                export const Foo = 1;
                export const Foo = 2;
              }
              export namespace Baz {
                export const Bar = 3;
                export const Bar = 4;
              }
            }
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'Foo'.",
          line: 4,
          column: 30,
          endLine: 4,
          endColumn: 33,
        },
        {
          message: "Multiple exports of name 'Foo'.",
          line: 5,
          column: 30,
          endLine: 5,
          endColumn: 33,
        },
        {
          message: "Multiple exports of name 'Bar'.",
          line: 8,
          column: 30,
          endLine: 8,
          endColumn: 33,
        },
        {
          message: "Multiple exports of name 'Bar'.",
          line: 9,
          column: 30,
          endLine: 9,
          endColumn: 33,
        },
      ],
    },
    {
      code: `
              export class Foo { }
              export class Foo { }
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'Foo'.",
          line: 2,
          column: 28,
          endLine: 2,
          endColumn: 31,
        },
        {
          message: "Multiple exports of name 'Foo'.",
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
              export enum Foo { }
              export enum Foo { }
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'Foo'.",
          line: 2,
          column: 27,
          endLine: 2,
          endColumn: 30,
        },
        {
          message: "Multiple exports of name 'Foo'.",
          line: 3,
          column: 27,
          endLine: 3,
          endColumn: 30,
        },
      ],
    },
    {
      code: `
              export enum Foo { }
              export class Foo { }
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'Foo'.",
          line: 2,
          column: 27,
          endLine: 2,
          endColumn: 30,
        },
        {
          message: "Multiple exports of name 'Foo'.",
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
              export const Foo = 'bar';
              export class Foo { }
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'Foo'.",
          line: 2,
          column: 28,
          endLine: 2,
          endColumn: 31,
        },
        {
          message: "Multiple exports of name 'Foo'.",
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
              export function Foo() { };
              export class Foo { }
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'Foo'.",
          line: 2,
          column: 31,
          endLine: 2,
          endColumn: 34,
        },
        {
          message: "Multiple exports of name 'Foo'.",
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
              export const Foo = 'bar';
              export function Foo() { };
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'Foo'.",
          line: 2,
          column: 28,
          endLine: 2,
          endColumn: 31,
        },
        {
          message: "Multiple exports of name 'Foo'.",
          line: 3,
          column: 31,
          endLine: 3,
          endColumn: 34,
        },
      ],
    },
    {
      code: `
              export const Foo = 'bar';
              export namespace Foo { }
            `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'Foo'.",
          line: 2,
          column: 28,
          endLine: 2,
          endColumn: 31,
        },
        {
          message: "Multiple exports of name 'Foo'.",
          line: 3,
          column: 32,
          endLine: 3,
          endColumn: 35,
        },
      ],
    },
    {
      code: `
            declare module "a" {
              const Foo = 1;
              export {Foo as default};
            }
            const Bar = 2;
            export {Bar as default};
            const Baz = 3;
            export {Baz as default};
          `,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: 'Multiple default exports.',
          line: 7,
          column: 28,
          endLine: 7,
          endColumn: 35,
        },
        {
          message: 'Multiple default exports.',
          line: 9,
          column: 28,
          endLine: 9,
          endColumn: 35,
        },
      ],
    },
    // Documentation
    {
      code: `export default class MyClass { /*...*/ } // Multiple default exports.

function makeClass() { return new MyClass(...arguments) }

export default makeClass // Multiple default exports.`,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: 'Multiple default exports.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
        {
          message: 'Multiple default exports.',
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 25,
        },
      ],
    },
    {
      code: `export const foo = function () { /*...*/ } // Multiple exports of name 'foo'.

function bar() { /*...*/ }
export { bar as foo } // Multiple exports of name 'foo'.`,
      filename: testFixturePath('export-rule/consumer.ts'),
      errors: [
        {
          message: "Multiple exports of name 'foo'.",
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 17,
        },
        {
          message: "Multiple exports of name 'foo'.",
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 20,
        },
      ],
    },
  ],
});
