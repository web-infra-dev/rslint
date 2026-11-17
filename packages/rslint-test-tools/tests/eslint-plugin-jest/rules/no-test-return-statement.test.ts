// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-test-return-statement', {
  valid: [
    ...(
      [
        'it("noop", function () {});',
        'test("noop", () => {});',
        'test("one", () => expect(1).toBe(1));',
        'test("empty")',
        `test("one", () => {
  expect(1).toBe(1);
});`,
        `it("one", function () {
  expect(1).toBe(1);
});`,
        `it("one", myTest);
function myTest() {
  expect(1).toBe(1);
}`,
        `it("one", () => expect(1).toBe(1));
function myHelper() {}`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'it("noop", function () {});',
        'test("noop", () => {});',
        'test("one", () => expect(1).toBe(1));',
        'test("empty")',
        `test("one", () => {
  expect(1).toBe(1);
});`,
        `it("one", function () {
  expect(1).toBe(1);
});`,
        `it("one", myTest);
function myTest() {
  expect(1).toBe(1);
}`,
        `it("one", () => expect(1).toBe(1));
function myHelper() {}`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-test-return-statement / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          `test("one", () => {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.skip("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.each\`\`("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.each()("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.only.each\`\`("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.only.each()("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it("one", myTest);
function myTest () {
  return expect(1).toBe(1);
}`,
          3,
        ],
      ] as const
    ).map(([code, line], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'noReturnValue',
          message: 'Jest tests should not return a value',
          line,
          column: 3,
          endLine: line,
          endColumn: 28,
        },
      ],
    })),
    ...(
      [
        [
          `test("one", () => {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.skip("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.each\`\`("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.each()("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.only.each\`\`("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it.only.each()("one", function () {
  return expect(1).toBe(1);
});`,
          2,
        ],
        [
          `it("one", myTest);
function myTest () {
  return expect(1).toBe(1);
}`,
          3,
        ],
      ] as const
    ).map(([code, line], index): TestCase => ({
      name: `upstream no-test-return-statement / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noReturnValue',
          message: 'Jest tests should not return a value',
          line,
          column: 3,
          endLine: line,
          endColumn: 28,
        },
      ],
    })),
  ],
});
