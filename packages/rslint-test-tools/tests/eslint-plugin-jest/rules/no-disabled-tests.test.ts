// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-disabled-tests', {
  valid: [
    ...(
      [
        'describe("foo", function () {})',
        'it("foo", function () {})',
        'describe.only("foo", function () {})',
        'it.only("foo", function () {})',
        'it.each("foo", () => {})',
        'it.concurrent("foo", function () {})',
        'test("foo", function () {})',
        'test.only("foo", function () {})',
        'test.concurrent("foo", function () {})',
        'describe[`${"skip"}`]("foo", function () {})',
        'it.todo("fill this later")',
        'var appliedSkip = describe.skip; appliedSkip.apply(describe)',
        'var calledSkip = it.skip; calledSkip.call(it)',
        '({ f: function () {} }).f()',
        '(a || b).f()',
        'itHappensToStartWithIt()',
        'testSomething()',
        'xitSomethingElse()',
        'xitiViewMap()',
        `
        import { pending } from "actions"

        test("foo", () => {
          expect(pending()).toEqual({})
        })
     \u0020`,
        `
        const { pending } = require("actions")

        test("foo", () => {
          expect(pending()).toEqual({})
        })
     \u0020`,
        `
        test("foo", () => {
          const pending = getPending()
          expect(pending()).toEqual({})
        })
     \u0020`,
        `
        test("foo", () => {
          expect(pending()).toEqual({})
        })

        function pending() {
          return {}
        }
     \u0020`,
        `
        import { test } from './test-utils';

        test('something');
     \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'describe("foo", function () {})',
        'it("foo", function () {})',
        'describe.only("foo", function () {})',
        'it.only("foo", function () {})',
        'it.each("foo", () => {})',
        'it.concurrent("foo", function () {})',
        'test("foo", function () {})',
        'test.only("foo", function () {})',
        'test.concurrent("foo", function () {})',
        'describe[`${"skip"}`]("foo", function () {})',
        'it.todo("fill this later")',
        'var appliedSkip = describe.skip; appliedSkip.apply(describe)',
        'var calledSkip = it.skip; calledSkip.call(it)',
        '({ f: function () {} }).f()',
        '(a || b).f()',
        'itHappensToStartWithIt()',
        'testSomething()',
        'xitSomethingElse()',
        'xitiViewMap()',
        `import { pending } from "actions"

test("foo", () => {
  expect(pending()).toEqual({})
})`,
        `const { pending } = require("actions")

test("foo", () => {
  expect(pending()).toEqual({})
})`,
        `test("foo", () => {
  const pending = getPending()
  expect(pending()).toEqual({})
})`,
        `test("foo", () => {
  expect(pending()).toEqual({})
})

function pending() {
  return {}
}`,
        `import { test } from './test-utils';

test('something');`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-disabled-tests / valid ${index}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        ['describe.skip("foo", function () {})', 37],
        ['describe.skip.each([1, 2, 3])("%s", (a, b) => {});', 50],
        ['xdescribe.each([1, 2, 3])("%s", (a, b) => {});', 46],
        ['describe[`skip`]("foo", function () {})', 40],
        ['describe["skip"]("foo", function () {})', 40],
        ['it.skip("foo", function () {})', 31],
        ['it["skip"]("foo", function () {})', 34],
        ['test.skip("foo", function () {})', 33],
        ['it.skip.each``("foo", function () {})', 38],
        ['test.skip.each``("foo", function () {})', 40],
        ['it.skip.each([])("foo", function () {})', 40],
        ['test.skip.each([])("foo", function () {})', 42],
        ['test["skip"]("foo", function () {})', 36],
        ['xdescribe("foo", function () {})', 33],
        ['xit("foo", function () {})', 27],
        ['xtest("foo", function () {})', 29],
        ['xit.each``("foo", function () {})', 34],
        ['xtest.each``("foo", function () {})', 36],
        ['xit.each([])("foo", function () {})', 36],
        ['xtest.each([])("foo", function () {})', 38],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['it("has title but no callback")', 32],
        ['test("has title but no callback")', 34],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 20}`,
      code,
      errors: [
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['it("contains a call to pending", function () { pending() })', 48],
        ['pending();', 1],
        [
          'describe("contains a call to pending", function () { pending() })',
          54,
        ],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `run 0 invalid ${index + 22}`,
      code,
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 9,
        },
      ],
    })),
    {
      name: 'run 0 invalid 25',
      code: `
        import { test } from '@jest/globals';

        test('something');
     \u0020`,
      errors: [
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 26,
        },
      ],
    },
    ...(
      [
        ['describe.skip("foo", function () {})', 37],
        ['describe.skip.each([1, 2, 3])("%s", (a, b) => {});', 50],
        ['xdescribe.each([1, 2, 3])("%s", (a, b) => {});', 46],
        ['describe[`skip`]("foo", function () {})', 40],
        ['describe["skip"]("foo", function () {})', 40],
        ['it.skip("foo", function () {})', 31],
        ['it["skip"]("foo", function () {})', 34],
        ['test.skip("foo", function () {})', 33],
        ['it.skip.each``("foo", function () {})', 38],
        ['test.skip.each``("foo", function () {})', 40],
        ['it.skip.each([])("foo", function () {})', 40],
        ['test.skip.each([])("foo", function () {})', 42],
        ['test["skip"]("foo", function () {})', 36],
        ['xdescribe("foo", function () {})', 33],
        ['xit("foo", function () {})', 27],
        ['xtest("foo", function () {})', 29],
        ['xit.each``("foo", function () {})', 34],
        ['xtest.each``("foo", function () {})', 36],
        ['xit.each([])("foo", function () {})', 36],
        ['xtest.each([])("foo", function () {})', 38],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-disabled-tests / invalid ${index}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['it("has title but no callback")', 32],
        ['test("has title but no callback")', 34],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-disabled-tests / invalid ${index + 20}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['it("contains a call to pending", function () { pending() })', 48],
        ['pending();', 1],
        [
          'describe("contains a call to pending", function () { pending() })',
          54,
        ],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `upstream no-disabled-tests / invalid ${index + 22}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 9,
        },
      ],
    })),
    {
      name: 'upstream no-disabled-tests / invalid 25',
      code: `import { test } from '@jest/globals';

test('something');`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
  ],
});
