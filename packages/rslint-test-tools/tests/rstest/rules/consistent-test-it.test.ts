import { RuleTester } from '../rule-tester';

// Mirrors the adapted upstream Go suite; framework-specific extras stay in Go.
new RuleTester().run('consistent-test-it', {} as never, {
  valid: [
    {
      code: 'test("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'test.only("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'test.skip("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'test.concurrent("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'xtest("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'test.each([])("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'test.each``("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'it("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'fit("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'xit("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'it.only("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'it.skip("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'it.concurrent("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'it.each([])("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'it.each``("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'test("foo")',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'test.only("foo")',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'test.skip("foo")',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'test.concurrent("foo")',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'xtest("foo")',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: '[1,2,3].forEach(() => { test("foo") })',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'it("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'it.only("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'it.skip("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'it.concurrent("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'xit("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: '[1,2,3].forEach(() => { it("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'test("foo");',
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'it("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'test("foo")',
    },
    {
      code: 'test("foo")',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'test("foo")',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'it("shows error", () => {\n  expect(true).toBe(false);\n        });',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'it("foo", function () {\n         expect(true).toBe(false);\n     })',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: " it('foo', () => {\n      expect(true).toBe(false);\n  });\n  function myTest() { if ('bar') {} }",
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'bench("foo", function () {\n        fibonacci(10);\n     })',
      options: [
        {
          fn: 'it',
        },
      ],
    },
    {
      code: 'test("shows error", () => {\n      expect(true).toBe(false);\n     });',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'test.skip("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'test.concurrent("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'xtest("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'test.each([])("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'test.each``("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'test',
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'it("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'test("shows error", () => {});',
    },
    {
      code: 'test("foo")',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
    },
    {
      code: 'test("foo")',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
    },
    {
      code: 'import { it as baseIt, test } from "@rstest/core"\nbaseIt("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
    },
  ],
  invalid: [
    {
      code: 'it("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'import { it } from \'@rstest/core\';\n\nit("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 3,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'import { it as testThisThing } from \'@rstest/core\';\n\ntestThisThing("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 3,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'it.skip("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'it.concurrent("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'it.only("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'it.each([])("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'it.each``("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'describe.each``("foo", () => { it.each``("bar") })',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 32,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'describe.each``("foo", () => { test.each``("bar") })',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 32,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'describe.each()("%s", () => {\n  test("is valid, but should not be", () => {});\n\n  it("is not valid, but should be", () => {});\n});',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 2,
          column: 3,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'describe.only.each()("%s", () => {\n  test("is valid, but should not be", () => {});\n\n  it("is not valid, but should be", () => {});\n});',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 2,
          column: 3,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'test("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'test.skip("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'test.concurrent("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'test.only("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'test.each([])("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'describe.each``("foo", () => { test.each``("bar") })',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 32,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'test.each``("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test.only("foo") })',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test.skip("foo") })',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test.concurrent("foo") })',
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { it.only("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { it.skip("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { it.concurrent("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'it("foo")',
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'test("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'it("foo")',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'it("foo")',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'test("shows error", () => {});',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'test.skip("shows error");',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: "test.only('shows error');",
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: "describe('foo', () => { it('bar', () => {}); });",
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 25,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'import { test } from "@rstest/core"\ntest("shows error", () => {});',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 2,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'import { expect, test, it } from "@rstest/core"\ntest("shows error", () => {});',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 2,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'it("shows error", () => {});',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'test("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'it' instead of 'test'",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'it("foo")',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'it' instead of 'test' within describe",
        },
      ],
    },
    {
      code: 'it("foo")',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 1,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'import { it } from "@rstest/core"\nit("foo")',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 2,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'import { expect, it, test } from "@rstest/core"\nit("foo")',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          line: 2,
          column: 1,
          message: "Prefer using 'test' instead of 'it'",
        },
      ],
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          line: 1,
          column: 27,
          message: "Prefer using 'test' instead of 'it' within describe",
        },
      ],
    },
  ],
});
