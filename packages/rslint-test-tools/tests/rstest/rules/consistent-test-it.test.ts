import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Mirrors the adapted upstream Go suite; framework-specific extras stay in Go.
new RuleTester(config).run('consistent-test-it', {
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 3,
        },
      ],
      output: 'import { test, it } from \'@rstest/core\';\n\ntest("foo")',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 14,
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
      output: 'test.skip("foo")',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
      output: 'test.concurrent("foo")',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
      output: 'test.only("foo")',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: 'test.each([])("foo")',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: 'test.each``("foo")',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 41,
        },
      ],
      output: 'describe.each``("foo", () => { test.each``("bar") })',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 43,
        },
      ],
      output: 'describe.each``("foo", () => { it.each``("bar") })',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 7,
        },
      ],
      output:
        'describe.each()("%s", () => {\n  it("is valid, but should not be", () => {});\n\n  it("is not valid, but should be", () => {});\n});',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 7,
        },
      ],
      output:
        'describe.only.each()("%s", () => {\n  it("is valid, but should not be", () => {});\n\n  it("is not valid, but should be", () => {});\n});',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
      output: 'it("foo")',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: 'it.skip("foo")',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: 'it.concurrent("foo")',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: 'it.only("foo")',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
      output: 'it.each([])("foo")',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 43,
        },
      ],
      output: 'describe.each``("foo", () => { it.each``("bar") })',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: 'it.each``("foo")',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 36,
        },
      ],
      output: 'describe("suite", () => { it.only("foo") })',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 36,
        },
      ],
      output: 'describe("suite", () => { it.skip("foo") })',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 42,
        },
      ],
      output: 'describe("suite", () => { it.concurrent("foo") })',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'describe("suite", () => { test.only("foo") })',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'describe("suite", () => { test.skip("foo") })',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 40,
        },
      ],
      output: 'describe("suite", () => { test.concurrent("foo") })',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
      output: 'it("foo")',
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
      output: 'it("shows error", () => {});',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: 'it.skip("shows error");',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: "it.only('shows error');",
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 27,
        },
      ],
      output: "describe('foo', () => { test('bar', () => {}); });",
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
          message: "Prefer using 'it' instead of 'test'",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 5,
        },
      ],
      output:
        'import { it, test } from "@rstest/core"\nit("shows error", () => {});',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 5,
        },
      ],
      output:
        'import { expect, test, it } from "@rstest/core"\nit("shows error", () => {});',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("shows error", () => {});',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
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
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
      output: 'it("foo")',
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
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
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 3,
        },
      ],
      output: 'import { test, it } from "@rstest/core"\ntest("foo")',
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
          message: "Prefer using 'test' instead of 'it'",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 3,
        },
      ],
      output: 'import { expect, it, test } from "@rstest/core"\ntest("foo")',
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
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
    },
  ],
});
