import { test } from '../utils.js';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();

const preferNamed = 'Prefer named exports.';
const noAliasDefault =
  'Do not alias `foo` as `default`. Just export `foo` itself instead.';

ruleTester.run('no-default-export', null as never, {
  valid: [
    test({ code: 'module.exports = function foo() {}' }),
    test({ code: 'module.exports = function foo() {}' }),
    test({
      code: `
        export const foo = 'foo';
        export const bar = 'bar';
      `,
    }),
    test({
      code: `
        export const foo = 'foo';
        export function bar() {};
      `,
    }),
    test({ code: "export const foo = 'foo';" }),
    test({
      code: `
        const foo = 'foo';
        export { foo };
      `,
    }),
    test({ code: 'let foo, bar; export { foo, bar }' }),
    test({ code: 'export const { foo, bar } = item;' }),
    test({ code: 'export const { foo, bar: baz } = item;' }),
    test({ code: 'export const { foo: { bar, baz } } = item;' }),
    test({
      code: `
        let item;
        export const foo = item;
        export { item };
      `,
    }),
    test({ code: "export * from './foo';" }),
    test({ code: 'export const { foo } = { foo: "bar" };' }),
    test({
      code: 'export const { foo: { bar } } = { foo: { bar: "baz" } };',
    }),
    test({ code: 'export { a, b } from "foo.js"' }),

    // no exports at all
    test({ code: "import * as foo from './foo';" }),
    test({ code: "import foo from './foo';" }),
    test({ code: "import {default as foo} from './foo';" }),

    test({ code: 'export type UserId = number;' }),
  ],
  invalid: [
    test({
      code: 'export default function bar() {};',
      errors: [{ message: preferNamed }],
    }),
    test({
      code: `
        export const foo = 'foo';
        export default bar;`,
      errors: [{ message: preferNamed }],
    }),
    test({
      code: 'export default class Bar {};',
      errors: [{ message: preferNamed }],
    }),
    test({
      code: 'export default function() {};',
      errors: [{ message: preferNamed }],
    }),
    test({
      code: 'export default class {};',
      errors: [{ message: preferNamed }],
    }),
    test({
      code: 'let foo; export { foo as default }',
      errors: [{ message: noAliasDefault }],
    }),
    test({
      code: 'let foo; export { foo as "default" }',
      errors: [{ message: noAliasDefault }],
    }),
  ],
});
