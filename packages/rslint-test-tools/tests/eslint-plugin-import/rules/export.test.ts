import { test, testFilePath } from '../utils.js';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
// Rslint-Modified: we don't require rules
// const rule = require('rules/export')
const rule = null as never;
// Rslint-Modified end

ruleTester.run('export', rule, {
  valid: [
    test({ code: 'import "./malformed.js"' }),
    test({ code: 'var foo = "foo"; export default foo;' }),
    test({ code: 'export var foo = "foo"; export var bar = "bar";' }),
    test({ code: 'export var foo = "foo", bar = "bar";' }),
    test({ code: 'export var { foo, bar } = object;' }),
    test({ code: 'export var [ foo, bar ] = array;' }),
    test({ code: 'let foo; export { foo, foo as bar }' }),
    test({ code: 'export * from "./does-not-exist"' }),
    test({ code: 'export default foo; export * from "./bar"' }),
    test({
      code: `
        export default function foo(param: string): boolean;
        export default function foo(param: string, param1?: number): boolean {
          return Boolean(param) && param1 !== undefined;
        }
      `,
    }),
    test({
      code: `
        export const Foo = 1;
        export type Foo = number;
      `,
    }),
    test({
      code: `
        export const Foo = 1;
        export interface Foo {}
      `,
    }),
    test({
      code: `
        export function fff(a: string): void;
        export function fff(a: number): void;
      `,
    }),
    test({
      code: `
        export const Bar = 1;
        export namespace Foo {
          export const Bar = 1;
        }
      `,
    }),
    test({
      code: `
        export class Foo {}
        export namespace Foo {}
        export namespace Foo {
          export class Bar {}
        }
      `,
    }),
    test({
      code: `
        export function Foo(): void;
        export namespace Foo {}
      `,
    }),
    test({
      code: `
        export enum Foo {}
        export namespace Foo {}
      `,
    }),
    test({
      code: `
        declare module "a" {
          const Foo = 1;
          export { Foo as default };
        }
        declare module "b" {
          const Bar = 2;
          export { Bar as default };
        }
      `,
    }),
    // export * from a module that resolves to a real file with named exports —
    // no duplicate when the local module has no overlapping names.
    test({
      code: 'export * from "./fixtures/export-all";',
      filename: testFilePath('./main.ts'),
    }),
  ],
  invalid: [
    test({
      code: `
        export type Foo = string;
        export type Foo = number;
      `,
      errors: [
        { message: "Multiple exports of name 'Foo'." },
        { message: "Multiple exports of name 'Foo'." },
      ],
    }),
    test({
      code: `
        export interface Foo {}
        export interface Foo {}
      `,
      errors: [
        { message: "Multiple exports of name 'Foo'." },
        { message: "Multiple exports of name 'Foo'." },
      ],
    }),
    test({
      code: `
        export const a = 1;
        export namespace Foo {
          export const a = 2;
          export const a = 3;
        }
      `,
      errors: [
        { message: "Multiple exports of name 'a'." },
        { message: "Multiple exports of name 'a'." },
      ],
    }),
    test({
      code: `
        declare module "foo" {
          const Foo = 1;
          export default Foo;
          export default Foo;
        }
      `,
      errors: [
        { message: 'Multiple default exports.' },
        { message: 'Multiple default exports.' },
      ],
    }),
    test({
      code: `
        export class Foo {}
        export class Foo {}
        export namespace Foo {}
      `,
      errors: [
        { message: "Multiple exports of name 'Foo'." },
        { message: "Multiple exports of name 'Foo'." },
      ],
    }),
    test({
      code: `
        export const Foo = 'bar';
        export namespace Foo {}
      `,
      errors: [
        { message: "Multiple exports of name 'Foo'." },
        { message: "Multiple exports of name 'Foo'." },
      ],
    }),
    test({
      code: `
        declare module "a" {
          const Foo = 1;
          export { Foo as default };
        }
        const Bar = 2;
        export { Bar as default };
        const Baz = 3;
        export { Baz as default };
      `,
      errors: [
        { message: 'Multiple default exports.' },
        { message: 'Multiple default exports.' },
      ],
    }),
    // export-all expansion: local `foo` collides with the `foo` re-exported
    // from `./fixtures/export-all`.
    test({
      code: 'export const foo = 1; export * from "./fixtures/export-all";',
      filename: testFilePath('./main.ts'),
      errors: [
        { message: "Multiple exports of name 'foo'." },
        { message: "Multiple exports of name 'foo'." },
      ],
    }),
  ],
});
