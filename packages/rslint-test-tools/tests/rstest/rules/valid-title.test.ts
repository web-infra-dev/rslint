import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-title', {} as never, {
  valid: [
    {
      code: 'test("does something", () => {});',
    },
    {
      code: 'describe("suite", () => { it("works", () => {}); });',
    },
    // Rstest expands %O and %c, unlike jest
    {
      code: 'test.each([])("%O and %c", () => {});',
    },
    {
      code: 'test.for([{ a: 1 }])("adds $a", () => {});',
    },
    {
      code: 'import { test } from "vitest";\ntest("test foo", () => {});',
    },
    {
      code: 'it("has a #unit tag", () => {});',
      options: [{ mustMatch: '#(?:unit|e2e)' }],
    },
  ],
  invalid: [
    {
      code: 'test("", () => {});',
      errors: [
        {
          messageId: 'emptyTitle',
          message: 'test should not have an empty title',
          line: 1,
          column: 1,
        },
      ],
    },
    {
      code: 'describe(123, () => {});',
      errors: [{ messageId: 'titleMustBeString', line: 1, column: 10 }],
    },
    {
      code: 'test(" foo", () => {});',
      output: 'test("foo", () => {});',
      errors: [{ messageId: 'accidentalSpace', line: 1, column: 6 }],
    },
    {
      code: 'test("test foo", () => {});',
      output: 'test("foo", () => {});',
      errors: [{ messageId: 'duplicatePrefix', line: 1, column: 6 }],
    },
    // %p is a jest placeholder that Rstest does not expand
    {
      code: 'test.each([])("%p", () => {});',
      errors: [
        {
          messageId: 'invalidEachSpecifier',
          message: '"%p" is not a valid format specifier',
        },
      ],
    },
    // .for formats its title the same way .each does
    {
      code: 'test.for([])("%z", () => {});',
      errors: [{ messageId: 'invalidEachSpecifier' }],
    },
    {
      code: 'test("the correct way", () => {});',
      options: [{ disallowedWords: ['correct'] }],
      errors: [{ messageId: 'disallowedWord' }],
    },
    {
      code: 'test("nope", () => {});',
      options: [{ mustMatch: '^ok' }],
      errors: [{ messageId: 'mustMatch', message: 'test should match /^ok/u' }],
    },
    {
      code: 'describe("nope", () => {});',
      options: [{ mustNotMatch: ['^nope', 'describe titles must be useful'] }],
      errors: [
        {
          messageId: 'mustNotMatchCustom',
          message: 'describe titles must be useful',
        },
      ],
    },
    {
      code: 'describe("nope", () => {});',
      options: [{ mustNotMatch: '^nope' }],
      errors: [
        {
          messageId: 'mustNotMatch',
          message: 'describe should not match /^nope/u',
        },
      ],
    },
    {
      code: 'test("nope", () => {});',
      options: [{ mustMatch: ['^ok', 'titles must start with "ok"'] }],
      errors: [
        {
          messageId: 'mustMatchCustom',
          message: 'titles must start with "ok"',
        },
      ],
    },
    // A pattern that does not compile is reported once, and no other diagnostic
    // is produced.
    {
      code: 'test("test foo", () => {});',
      options: [{ mustMatch: '(unclosed' }],
      errors: [{ messageId: 'invalidPattern', line: 1, column: 1 }],
    },
  ],
});
