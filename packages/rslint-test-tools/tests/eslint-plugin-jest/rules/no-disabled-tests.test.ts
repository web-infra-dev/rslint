import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-disabled-tests', {} as never, {
  valid: [
    { code: 'describe("foo", function () {})' },
    { code: 'it("foo", function () {})' },
    { code: 'describe.only("foo", function () {})' },
    { code: 'it.only("foo", function () {})' },
    { code: 'it.each("foo", () => {})' },
    { code: 'it.concurrent("foo", function () {})' },
    { code: 'test("foo", function () {})' },
    { code: 'test.only("foo", function () {})' },
    { code: 'test.concurrent("foo", function () {})' },
    { code: 'describe[`${"skip"}`]("foo", function () {})' },
    { code: 'it.todo("fill this later")' },
    { code: 'var appliedSkip = describe.skip; appliedSkip.apply(describe)' },
    { code: 'var calledSkip = it.skip; calledSkip.call(it)' },
    { code: '({ f: function () {} }).f()' },
    { code: '(a || b).f()' },
    { code: 'itHappensToStartWithIt()' },
    { code: 'testSomething()' },
    { code: 'xitSomethingElse()' },
    { code: 'xitiViewMap()' },
    {
      code: `
        import { pending } from "actions"

        test("foo", () => {
          expect(pending()).toEqual({})
        })
      `,
    },
    {
      code: `
        const { pending } = require("actions")

        test("foo", () => {
          expect(pending()).toEqual({})
        })
      `,
    },
    {
      code: `
        test("foo", () => {
          const pending = getPending()
          expect(pending()).toEqual({})
        })
      `,
    },
    {
      code: `
        test("foo", () => {
          expect(pending()).toEqual({})
        })

        function pending() {
          return {}
        }
      `,
    },
    {
      code: `
        import { test } from './test-utils';

        test('something');
      `,
    },
  ],
  invalid: [
    {
      code: 'describe.skip("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'describe.skip.each([1, 2, 3])("%s", (a, b) => {});',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'xdescribe.each([1, 2, 3])("%s", (a, b) => {});',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'describe[`skip`]("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'describe["skip"]("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'it.skip("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'it["skip"]("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'test.skip("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'it.skip.each``("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'test.skip.each``("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'it.skip.each([])("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'test.skip.each([])("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'test["skip"]("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'xdescribe("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'xit("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'xtest("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'xit.each``("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'xtest.each``("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'xit.each([])("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'xtest.each([])("foo", function () {})',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'it("has title but no callback")',
      errors: [{ message: 'Test is missing function argument' }],
    },
    {
      code: 'test("has title but no callback")',
      errors: [{ message: 'Test is missing function argument' }],
    },
    {
      code: 'it("contains a call to pending", function () { pending() })',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'pending();',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: 'describe("contains a call to pending", function () { pending() })',
      errors: [{ message: 'Tests should not be skipped' }],
    },
    {
      code: `
        import { test } from '@jest/globals';

        test('something');
      `,
      errors: [{ message: 'Test is missing function argument' }],
    },
  ],
});
