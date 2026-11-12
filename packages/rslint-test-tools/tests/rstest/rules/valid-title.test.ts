import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('valid-title', {
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
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'describe(123, () => {});',
      errors: [
        {
          messageId: 'titleMustBeString',
          message: 'Title must be a string',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'test(" foo", () => {});',
      output: 'test("foo", () => {});',
      errors: [
        {
          messageId: 'accidentalSpace',
          message: 'should not have leading or trailing spaces',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'test("test foo", () => {});',
      output: 'test("foo", () => {});',
      errors: [
        {
          messageId: 'duplicatePrefix',
          message: 'should not have duplicate prefix',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    // %p is a jest placeholder that Rstest does not expand
    {
      code: 'test.each([])("%p", () => {});',
      errors: [
        {
          messageId: 'invalidEachSpecifier',
          message: '"%p" is not a valid format specifier',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    // .for formats its title the same way .each does
    {
      code: 'test.for([])("%z", () => {});',
      errors: [
        {
          messageId: 'invalidEachSpecifier',
          message: '"%z" is not a valid format specifier',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'test("the correct way", () => {});',
      options: [{ disallowedWords: ['correct'] }],
      errors: [
        {
          messageId: 'disallowedWord',
          message: '"correct" is not allowed in test titles',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'test("nope", () => {});',
      options: [{ mustMatch: '^ok' }],
      errors: [
        {
          messageId: 'mustMatch',
          message: 'test should match /^ok/u',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'describe("nope", () => {});',
      options: [{ mustNotMatch: ['^nope', 'describe titles must be useful'] }],
      errors: [
        {
          messageId: 'mustNotMatchCustom',
          message: 'describe titles must be useful',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
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
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
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
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    // A pattern that does not compile is reported once, and no other diagnostic
    // is produced.
    {
      code: 'test("test foo", () => {});',
      options: [{ mustMatch: '(unclosed' }],
      errors: [
        {
          messageId: 'invalidPattern',
          message:
            'Invalid regular expression in `mustMatch` option: `(unclosed`: error parsing regexp: missing closing ) in `(unclosed`',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
    },
  ],
});
