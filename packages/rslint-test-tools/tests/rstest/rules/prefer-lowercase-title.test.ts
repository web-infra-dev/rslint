import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-lowercase-title', {
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
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`test`s should begin with lowercase',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: `test("Foo", function () {})`,
      output: `test("foo", function () {})`,
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`test`s should begin with lowercase',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'test(`Foo`, function () {})',
      output: 'test(`foo`, function () {})',
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`test`s should begin with lowercase',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: `describe('Foo', function () {})`,
      output: `describe('foo', function () {})`,
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`describe`s should begin with lowercase',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: `describe("Foo", function () {})`,
      output: `describe("foo", function () {})`,
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`describe`s should begin with lowercase',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: `it('Foo', function () {})`,
      output: `it('foo', function () {})`,
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`it`s should begin with lowercase',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: "test('Doesn\\'t mutate', () => {})",
      output: "test('doesn\\'t mutate', () => {})",
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`test`s should begin with lowercase',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: `const todoTest = test.todo; todoTest('Should work');`,
      output: `const todoTest = test.todo; todoTest('should work');`,
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`test`s should begin with lowercase',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 51,
        },
      ],
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
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`describe`s should begin with lowercase',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 19,
        },
      ],
    },
  ],
});
