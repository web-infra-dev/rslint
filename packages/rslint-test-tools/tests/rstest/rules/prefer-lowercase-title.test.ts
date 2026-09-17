import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-lowercase-title', {} as never, {
  valid: [
    { code: 'test()' },
    { code: `test('foo', function () {})` },
    { code: `test("foo", function () {})` },
    { code: 'test(`foo`, function () {})' },
    { code: 'test(42)' },
    { code: `test("")` },
    { code: 'describe()' },
    { code: `describe('foo', function () {})` },
    { code: `describe("foo", function () {})` },
    { code: 'describe(`foo`, function () {})' },
    { code: 'describe(42)' },
    { code: `describe("")` },
    { code: `it('foo', function () {})` },
    { code: 'it()' },
    { code: 'randomFunction()' },
    { code: 'foo.bar()' },
  ],
  invalid: [
    {
      code: `test('Foo', function () {})`,
      output: `test('foo', function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 6 }],
    },
    {
      code: `test("Foo", function () {})`,
      output: `test("foo", function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 6 }],
    },
    {
      code: 'test(`Foo`, function () {})',
      output: 'test(`foo`, function () {})',
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 6 }],
    },
    {
      code: `describe('Foo', function () {})`,
      output: `describe('foo', function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 10 }],
    },
    {
      code: `describe("Foo", function () {})`,
      output: `describe("foo", function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 10 }],
    },
    {
      code: `it('Foo', function () {})`,
      output: `it('foo', function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 4 }],
    },
  ],
});
