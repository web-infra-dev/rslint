import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-commented-out-tests', {} as never, {
  valid: [
    { code: '// foo("bar", function () {})' },
    { code: 'describe("foo", function () {})' },
    { code: 'it("foo", function () {})' },
    { code: 'describe.only("foo", function () {})' },
    { code: 'it.only("foo", function () {})' },
    { code: 'it.concurrent("foo", function () {})' },
    { code: 'test("foo", function () {})' },
    { code: 'test.only("foo", function () {})' },
    { code: 'test.concurrent("foo", function () {})' },
    { code: 'var appliedSkip = describe.skip; appliedSkip.apply(describe)' },
    { code: 'var calledSkip = it.skip; calledSkip.call(it)' },
    { code: '({ f: function () {} }).f()' },
    { code: '(a || b).f()' },
    { code: 'itHappensToStartWithIt()' },
    { code: 'testSomething()' },
    { code: '// latest(dates)' },
    { code: '// TODO: unify with Git implementation from Shipit (?)' },
    { code: '#!/usr/bin/env node' },
    { code: `
      import { pending } from "actions"

      test("foo", () => {
        expect(pending()).toEqual({})
      })
    ` },
    { code: `
      const { pending } = require("actions")

      test("foo", () => {
        expect(pending()).toEqual({})
      })
    ` },
    { code: `
      test("foo", () => {
        const pending = getPending()
        expect(pending()).toEqual({})
      })
    ` },
    { code: `
      test("foo", () => {
        expect(pending()).toEqual({})
      })

      function pending() {
        return {}
      }
    ` },
  ],

  invalid: [
    {
      code: '// describe("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// describe["skip"]("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// describe[\'skip\']("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// it.skip("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// it.only("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// it.concurrent("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// it["skip"]("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// test.skip("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// test.concurrent("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// test["skip"]("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// xdescribe("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// xit("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// fit("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// xtest("foo", function () {})',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: `
        // test(
        //   "foo", function () {}
        // )
      `,
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: `
        /* test
          (
            "foo", function () {}
          )
        */
      `,
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// it("has title but no callback")',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// it()',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// test.someNewMethodThatMightBeAddedInTheFuture()',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// test["someNewMethodThatMightBeAddedInTheFuture"]()',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: '// test("has title but no callback")',
      errors: [{ messageId: 'commentedTests', column: 1, line: 1 }],
    },
    {
      code: `
        foo()
        /*
          describe("has title but no callback", () => {})
        */
        bar()
      `,
      errors: [{ messageId: 'commentedTests', column: 1, line: 2 }],
    },
  ],
});
