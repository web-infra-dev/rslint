// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-commented-out-tests', {
  valid: [
    ...(
      [
        '// foo("bar", function () {})',
        'describe("foo", function () {})',
        'it("foo", function () {})',
        'describe.only("foo", function () {})',
        'it.only("foo", function () {})',
        'it.concurrent("foo", function () {})',
        'test("foo", function () {})',
        'test.only("foo", function () {})',
        'test.concurrent("foo", function () {})',
        'var appliedSkip = describe.skip; appliedSkip.apply(describe)',
        'var calledSkip = it.skip; calledSkip.call(it)',
        '({ f: function () {} }).f()',
        '(a || b).f()',
        'itHappensToStartWithIt()',
        'testSomething()',
        '// latest(dates)',
        '// TODO: unify with Git implementation from Ship it (?)',
        '#!/usr/bin/env node',
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
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        '// foo("bar", function () {})',
        'describe("foo", function () {})',
        'it("foo", function () {})',
        'describe.only("foo", function () {})',
        'it.only("foo", function () {})',
        'it.concurrent("foo", function () {})',
        'test("foo", function () {})',
        'test.only("foo", function () {})',
        'test.concurrent("foo", function () {})',
        'var appliedSkip = describe.skip; appliedSkip.apply(describe)',
        'var calledSkip = it.skip; calledSkip.call(it)',
        '({ f: function () {} }).f()',
        '(a || b).f()',
        'itHappensToStartWithIt()',
        'testSomething()',
        '// latest(dates)',
        '// TODO: unify with Git implementation from Shipit (?)',
        '#!/usr/bin/env node',
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
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-commented-out-tests / valid ${index}`,
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
        ['// describe("foo", function () {})', 1, 1, 1, 35],
        ['// describe["skip"]("foo", function () {})', 1, 1, 1, 43],
        ['// describe[\'skip\']("foo", function () {})', 1, 1, 1, 43],
        ['// it.skip("foo", function () {})', 1, 1, 1, 34],
        ['// it.only("foo", function () {})', 1, 1, 1, 34],
        ['// it.concurrent("foo", function () {})', 1, 1, 1, 40],
        ['// it["skip"]("foo", function () {})', 1, 1, 1, 37],
        ['// test.skip("foo", function () {})', 1, 1, 1, 36],
        ['// test.concurrent("foo", function () {})', 1, 1, 1, 42],
        ['// test["skip"]("foo", function () {})', 1, 1, 1, 39],
        ['// xdescribe("foo", function () {})', 1, 1, 1, 36],
        ['// xit("foo", function () {})', 1, 1, 1, 30],
        ['// fit("foo", function () {})', 1, 1, 1, 30],
        ['// xtest("foo", function () {})', 1, 1, 1, 32],
        [
          `
        // test(
        //   "foo", function () {}
        // )
     \u0020`,
          2,
          9,
          2,
          17,
        ],
        [
          `
        /* test
          (
            "foo", function () {}
          )
        */
     \u0020`,
          2,
          9,
          6,
          11,
        ],
        ['// it("has title but no callback")', 1, 1, 1, 35],
        ['// it()', 1, 1, 1, 8],
        ['// test.someNewMethodThatMightBeAddedInTheFuture()', 1, 1, 1, 51],
        ['// test["someNewMethodThatMightBeAddedInTheFuture"]()', 1, 1, 1, 54],
        ['// test("has title but no callback")', 1, 1, 1, 37],
        [
          `
        foo()
        /*
          describe("has title but no callback", () => {})
        */
        bar()
     \u0020`,
          3,
          9,
          5,
          11,
        ],
      ] as const
    ).map(([code, line, column, endLine, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line,
          column,
          endLine,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['// describe("foo", function () {})', 1, 1, 35],
        ['// describe["skip"]("foo", function () {})', 1, 1, 43],
        ['// describe[\'skip\']("foo", function () {})', 1, 1, 43],
        ['// it.skip("foo", function () {})', 1, 1, 34],
        ['// it.only("foo", function () {})', 1, 1, 34],
        ['// it.concurrent("foo", function () {})', 1, 1, 40],
        ['// it["skip"]("foo", function () {})', 1, 1, 37],
        ['// test.skip("foo", function () {})', 1, 1, 36],
        ['// test.concurrent("foo", function () {})', 1, 1, 42],
        ['// test["skip"]("foo", function () {})', 1, 1, 39],
        ['// xdescribe("foo", function () {})', 1, 1, 36],
        ['// xit("foo", function () {})', 1, 1, 30],
        ['// fit("foo", function () {})', 1, 1, 30],
        ['// xtest("foo", function () {})', 1, 1, 32],
        [
          `// test(
//   "foo", function () {}
// )`,
          1,
          1,
          9,
        ],
        [
          `/* test
  (
    "foo", function () {}
  )
*/`,
          1,
          5,
          3,
        ],
        ['// it("has title but no callback")', 1, 1, 35],
        ['// it()', 1, 1, 8],
        ['// test.someNewMethodThatMightBeAddedInTheFuture()', 1, 1, 51],
        ['// test["someNewMethodThatMightBeAddedInTheFuture"]()', 1, 1, 54],
        ['// test("has title but no callback")', 1, 1, 37],
        [
          `foo()
/*
  describe("has title but no callback", () => {})
*/
bar()`,
          2,
          4,
          3,
        ],
      ] as const
    ).map(([code, line, endLine, endColumn], index): TestCase => ({
      name: `upstream no-commented-out-tests / invalid ${index}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line,
          column: 1,
          endLine,
          endColumn,
        },
      ],
    })),
  ],
});
