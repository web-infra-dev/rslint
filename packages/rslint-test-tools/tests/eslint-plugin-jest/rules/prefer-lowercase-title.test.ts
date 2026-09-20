import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-lowercase-title', {} as never, {
  valid: [
    { code: 'it.each()' },
    { code: 'it.each()(1)' },
    { code: 'randomFunction()' },
    { code: 'foo.bar()' },
    { code: 'it()' },
    { code: `it(' ', function () {})` },
    { code: 'it(true, function () {})' },
    { code: 'it(MY_CONSTANT, function () {})' },
    { code: `it(" ", function () {})` },
    { code: 'it(` `, function () {})' },
    { code: `it('foo', function () {})` },
    { code: `it("foo", function () {})` },
    { code: 'it(`foo`, function () {})' },
    { code: `it("<Foo/>", function () {})` },
    { code: `it("123 foo", function () {})` },
    { code: 'it(42, function () {})' },
    { code: 'it(``)' },
    { code: `it("")` },
    { code: 'it(42)' },
    { code: 'test()' },
    { code: `test('foo', function () {})` },
    { code: `test("foo", function () {})` },
    { code: 'test(`foo`, function () {})' },
    { code: `test("<Foo/>", function () {})` },
    { code: `test("123 foo", function () {})` },
    { code: `test("42", function () {})` },
    { code: 'test(``)' },
    { code: `test("")` },
    { code: 'test(42)' },
    { code: 'describe()' },
    { code: `describe('foo', function () {})` },
    { code: `describe("foo", function () {})` },
    { code: 'describe(`foo`, function () {})' },
    { code: `describe("<Foo/>", function () {})` },
    { code: `describe("123 foo", function () {})` },
    { code: `describe("42", function () {})` },
    { code: 'describe(function () {})' },
    { code: 'describe(``)' },
    { code: `describe("")` },
    { code: 'describe.each()(1); describe.each()(2);' },
    { code: 'jest.doMock("my-module")' },
    {
      code: `
        import { jest } from '@jest/globals';
        jest.doMock('my-module');
      `,
    },
    { code: 'describe(42)' },
  ],
  invalid: [
    {
      code: `it('Foo', function () {})`,
      output: `it('foo', function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 4 }],
    },
    {
      code: `xit('Foo', function () {})`,
      output: `xit('foo', function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 5 }],
    },
    {
      code: `it("Foo", function () {})`,
      output: `it("foo", function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 4 }],
    },
    {
      code: 'it(`Foo`, function () {})',
      output: 'it(`foo`, function () {})',
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 4 }],
    },
    {
      code: `test('Foo', function () {})`,
      output: `test('foo', function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 6 }],
    },
    {
      code: `xtest('Foo', function () {})`,
      output: `xtest('foo', function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 7 }],
    },
    {
      code: `describe('Foo', function () {})`,
      output: `describe('foo', function () {})`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 10 }],
    },
    {
      code: 'describe(`Foo`, function () {})',
      output: 'describe(`foo`, function () {})',
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 10 }],
    },
    {
      code: "it.each(['green', 'black'])('Should return %', () => {})",
      output: "it.each(['green', 'black'])('should return %', () => {})",
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 29 }],
    },
    {
      code: "test('Doesn\\'t mutate', () => {})",
      output: "test('doesn\\'t mutate', () => {})",
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 6 }],
    },
    {
      code: 'test(`Value: \\${name}`, () => {})',
      output: 'test(`value: \\${name}`, () => {})',
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 6 }],
    },
  ],
});
