import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-export', {} as never, {
  valid: [
    { code: 'describe("a test", () => { expect(1).toBe(1); })' },
    { code: 'window.location = "valid"' },
    { code: 'module.somethingElse = "foo";' },
    { code: 'export const myThing = "valid"' },
    { code: 'export default function () {}' },
    { code: 'module.exports = function(){}' },
    { code: 'module.exports.myThing = "valid";' },
    { code: 'module.export.foo = "valid"; test("a test", () => {});' },
    {
      code: 'const exports = "exports"; module[exports] = {}; test("a test", () => {});',
    },
    {
      code: 'const module = { exports: {} }; module.exports.foo = "valid"; test("a test", () => {});',
    },
    {
      code: 'const run = (module: { exports: object }) => { module.exports.foo = "valid" }; test("a test", () => {});',
    },
    {
      code: 'const exports = { foo: "" }; exports.foo = "valid"; test("a test", () => {});',
    },
    {
      code: 'const run = (exports: { foo: string }) => { exports.foo = "valid" }; test("a test", () => {});',
    },
  ],
  invalid: [
    {
      code: 'export const myThing = "invalid"; test("a test", () => { expect(1).toBe(1);});',
      errors: [{ endColumn: 34, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: `
        export const myThing = 'invalid';

        test.each()('my code', () => {
          expect(1).toBe(1);
        });
      `,
      errors: [{ endColumn: 34, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: `
        export const myThing = 'invalid';

        test.each\`\`('my code', () => {
          expect(1).toBe(1);
        });
      `,
      errors: [{ endColumn: 34, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: `
        export const myThing = 'invalid';

        test.only.each\`\`('my code', () => {
          expect(1).toBe(1);
        });
      `,
      errors: [{ endColumn: 34, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'export default function() {};  test("a test", () => { expect(1).toBe(1);});',
      errors: [{ endColumn: 29, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'export = function() {}; test("a test", () => { expect(1).toBe(1);});',
      errors: [{ endColumn: 24, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'module.exports["invalid"] = function() {};  test("a test", () => { expect(1).toBe(1);});',
      errors: [{ endColumn: 26, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'module.exports = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
      errors: [{ endColumn: 15, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'module["exports"] = function() {}; test("a test", () => {});',
      errors: [{ endColumn: 18, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'module[`exports`].foo = function() {}; test("a test", () => {});',
      errors: [{ endColumn: 22, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'module.exports.foo.bar = function() {}; test("a test", () => {});',
      errors: [{ endColumn: 23, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'module.exports ||= {}; test("a test", () => {});',
      errors: [{ endColumn: 15, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'value = module.exports; test("a test", () => {});',
      errors: [{ endColumn: 23, column: 9, messageId: 'unexpectedExport' }],
    },
    {
      code: 'exports.foo = "invalid"; test("a test", () => {});',
      errors: [{ endColumn: 12, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'exports["foo"].bar = "invalid"; test("a test", () => {});',
      errors: [{ endColumn: 19, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'export import foo = require("./foo"); test("a test", () => {});',
      errors: [{ endColumn: 38, column: 1, messageId: 'unexpectedExport' }],
    },
    {
      code: 'export const myThing = "invalid"; describe("a suite", () => {});',
      errors: [{ endColumn: 34, column: 1, messageId: 'unexpectedExport' }],
    },
  ],
});
