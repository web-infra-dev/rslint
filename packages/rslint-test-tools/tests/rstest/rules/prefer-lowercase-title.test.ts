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
    {
      code: `const todoTest = test.todo; todoTest('Should work');`,
      options: [{ ignoreTodos: true }],
    },
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
    {
      code: "test('Doesn\\'t mutate', () => {})",
      output: "test('doesn\\'t mutate', () => {})",
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 6 }],
    },
    {
      code: `const todoTest = test.todo; todoTest('Should work');`,
      output: `const todoTest = test.todo; todoTest('should work');`,
      errors: [{ messageId: 'unexpectedCase', line: 1, column: 38 }],
    },
    {
      code: `describe('Outer', suiteBody);

function suiteBody() {
  describe('Inner', () => {});
}`,
      output: `describe('Outer', suiteBody);

function suiteBody() {
  describe('inner', () => {});
}`,
      options: [{ ignoreTopLevelDescribe: true }],
      errors: [{ messageId: 'unexpectedCase', line: 4, column: 12 }],
    },
  ],
});
