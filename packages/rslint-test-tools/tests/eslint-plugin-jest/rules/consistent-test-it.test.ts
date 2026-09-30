import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('consistent-test-it', {} as never, {
  valid: [
    { code: 'test("foo")', options: [{ fn: 'test' }] },
    { code: 'test.only("foo")', options: [{ fn: 'test' }] },
    { code: 'test.skip("foo")', options: [{ fn: 'test' }] },
    { code: 'test.concurrent("foo")', options: [{ fn: 'test' }] },
    { code: 'xtest("foo")', options: [{ fn: 'test' }] },
    { code: 'test.each([])("foo")', options: [{ fn: 'test' }] },
    { code: 'test.each``("foo")', options: [{ fn: 'test' }] },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [{ fn: 'test' }],
    },
    { code: 'it("foo")', options: [{ fn: 'it' }] },
    { code: 'fit("foo")', options: [{ fn: 'it' }] },
    { code: 'xit("foo")', options: [{ fn: 'it' }] },
    { code: 'it.only("foo")', options: [{ fn: 'it' }] },
    { code: 'it.skip("foo")', options: [{ fn: 'it' }] },
    { code: 'it.concurrent("foo")', options: [{ fn: 'it' }] },
    { code: 'it.each([])("foo")', options: [{ fn: 'it' }] },
    { code: 'it.each``("foo")', options: [{ fn: 'it' }] },
    { code: 'describe("suite", () => { it("foo") })', options: [{ fn: 'it' }] },
    { code: 'test("foo")', options: [{ fn: 'test', withinDescribe: 'it' }] },
    {
      code: 'test.only("foo")',
      options: [{ fn: 'test', withinDescribe: 'it' }],
    },
    {
      code: 'test.skip("foo")',
      options: [{ fn: 'test', withinDescribe: 'it' }],
    },
    {
      code: 'test.concurrent("foo")',
      options: [{ fn: 'test', withinDescribe: 'it' }],
    },
    { code: 'xtest("foo")', options: [{ fn: 'test', withinDescribe: 'it' }] },
    {
      code: '[1,2,3].forEach(() => { test("foo") })',
      options: [{ fn: 'test', withinDescribe: 'it' }],
    },
    { code: 'it("foo")', options: [{ fn: 'it', withinDescribe: 'test' }] },
    { code: 'it.only("foo")', options: [{ fn: 'it', withinDescribe: 'test' }] },
    { code: 'it.skip("foo")', options: [{ fn: 'it', withinDescribe: 'test' }] },
    {
      code: 'it.concurrent("foo")',
      options: [{ fn: 'it', withinDescribe: 'test' }],
    },
    { code: 'xit("foo")', options: [{ fn: 'it', withinDescribe: 'test' }] },
    {
      code: '[1,2,3].forEach(() => { it("foo") })',
      options: [{ fn: 'it', withinDescribe: 'test' }],
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [{ fn: 'test', withinDescribe: 'test' }],
    },
    { code: 'test("foo");', options: [{ fn: 'test', withinDescribe: 'test' }] },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [{ fn: 'it', withinDescribe: 'it' }],
    },
    { code: 'it("foo")', options: [{ fn: 'it', withinDescribe: 'it' }] },
    { code: 'test("foo")' },
    { code: 'test("foo")', options: [{ withinDescribe: 'it' }] },
    {
      code: 'describe("suite", () => { it("foo") })',
      options: [{ withinDescribe: 'it' }],
    },
    { code: 'test("foo")', options: [{ withinDescribe: 'test' }] },
    {
      code: 'describe("suite", () => { test("foo") })',
      options: [{ withinDescribe: 'test' }],
    },
  ],
  invalid: [
    {
      code: 'it("foo")',
      output: 'test("foo")',
      options: [{ fn: 'test' }],
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
    },
    {
      code: 'import { it } from \'@jest/globals\';\n\nit("foo")',
      // Differs from upstream, which calls the new name without importing it.
      output: 'import { test, it } from \'@jest/globals\';\n\ntest("foo")',
      options: [{ fn: 'test' }],
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
    },
    {
      code: 'import { it as testThisThing } from \'@jest/globals\';\n\ntestThisThing("foo")',
      // Differs from upstream, which calls the new name without importing it.
      output:
        'import { test, it as testThisThing } from \'@jest/globals\';\n\ntest("foo")',
      options: [{ fn: 'test' }],
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
      code: 'xit("foo")',
      output: 'xtest("foo")',
      options: [{ fn: 'test' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: 'fit("foo")',
      output: 'test.only("foo")',
      options: [{ fn: 'test' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: 'it.skip("foo")',
      output: 'test.skip("foo")',
      options: [{ fn: 'test' }],
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
    },
    {
      code: 'it.concurrent("foo")',
      output: 'test.concurrent("foo")',
      options: [{ fn: 'test' }],
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
    },
    {
      code: 'it.only("foo")',
      output: 'test.only("foo")',
      options: [{ fn: 'test' }],
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
    },
    {
      code: 'it.each([])("foo")',
      output: 'test.each([])("foo")',
      options: [{ fn: 'test' }],
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
    },
    {
      code: 'it.each``("foo")',
      output: 'test.each``("foo")',
      options: [{ fn: 'test' }],
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
    },
    {
      code: 'describe.each``("foo", () => { it.each``("bar") })',
      output: 'describe.each``("foo", () => { test.each``("bar") })',
      options: [{ fn: 'test' }],
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
    },
    {
      code: 'describe.each``("foo", () => { test.each``("bar") })',
      output: 'describe.each``("foo", () => { it.each``("bar") })',
      options: [{ fn: 'it' }],
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
    },
    {
      code: 'describe.each()("%s", () => {\n  test("is valid, but should not be", () => {});\n\n  it("is not valid, but should be", () => {});\n});',
      output:
        'describe.each()("%s", () => {\n  it("is valid, but should not be", () => {});\n\n  it("is not valid, but should be", () => {});\n});',
      options: [{ fn: 'test', withinDescribe: 'it' }],
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
    },
    {
      code: 'describe.only.each()("%s", () => {\n  test("is valid, but should not be", () => {});\n\n  it("is not valid, but should be", () => {});\n});',
      output:
        'describe.only.each()("%s", () => {\n  it("is valid, but should not be", () => {});\n\n  it("is not valid, but should be", () => {});\n});',
      options: [{ fn: 'test', withinDescribe: 'it' }],
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
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      output: 'describe("suite", () => { test("foo") })',
      options: [{ fn: 'test' }],
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
    },
    {
      code: 'test("foo")',
      output: 'it("foo")',
      options: [{ fn: 'it' }],
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
    },
    {
      code: 'xtest("foo")',
      output: 'xit("foo")',
      options: [{ fn: 'it' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'test.skip("foo")',
      output: 'it.skip("foo")',
      options: [{ fn: 'it' }],
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
    },
    {
      code: 'test.concurrent("foo")',
      output: 'it.concurrent("foo")',
      options: [{ fn: 'it' }],
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
    },
    {
      code: 'test.only("foo")',
      output: 'it.only("foo")',
      options: [{ fn: 'it' }],
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
    },
    {
      code: 'test.each([])("foo")',
      output: 'it.each([])("foo")',
      options: [{ fn: 'it' }],
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
    },
    {
      code: 'describe.each``("foo", () => { test.each``("bar") })',
      output: 'describe.each``("foo", () => { it.each``("bar") })',
      options: [{ fn: 'it' }],
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
    },
    {
      code: 'test.each``("foo")',
      output: 'it.each``("foo")',
      options: [{ fn: 'it' }],
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
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      output: 'describe("suite", () => { it("foo") })',
      options: [{ fn: 'it' }],
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
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      output: 'describe("suite", () => { it("foo") })',
      options: [{ fn: 'test', withinDescribe: 'it' }],
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
    },
    {
      code: 'describe("suite", () => { test.only("foo") })',
      output: 'describe("suite", () => { it.only("foo") })',
      options: [{ fn: 'test', withinDescribe: 'it' }],
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
    },
    {
      code: 'describe("suite", () => { xtest("foo") })',
      output: 'describe("suite", () => { xit("foo") })',
      options: [{ fn: 'test', withinDescribe: 'it' }],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'import { xtest as dontTestThis } from \'@jest/globals\';\n\ndescribe("suite", () => { dontTestThis("foo") });',
      // Differs from upstream, which calls the new name without importing it.
      output:
        'import { xit, xtest as dontTestThis } from \'@jest/globals\';\n\ndescribe("suite", () => { xit("foo") });',
      options: [{ fn: 'test', withinDescribe: 'it' }],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 3,
          column: 27,
          endLine: 3,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'import { describe as context, xtest as dontTestThis } from \'@jest/globals\';\n\ncontext("suite", () => { dontTestThis("foo") });',
      // Differs from upstream, which calls the new name without importing it.
      output:
        'import { describe as context, xit, xtest as dontTestThis } from \'@jest/globals\';\n\ncontext("suite", () => { xit("foo") });',
      options: [{ fn: 'test', withinDescribe: 'it' }],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 3,
          column: 26,
          endLine: 3,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'describe("suite", () => { test.skip("foo") })',
      output: 'describe("suite", () => { it.skip("foo") })',
      options: [{ fn: 'test', withinDescribe: 'it' }],
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
    },
    {
      code: 'describe("suite", () => { test.concurrent("foo") })',
      output: 'describe("suite", () => { it.concurrent("foo") })',
      options: [{ fn: 'test', withinDescribe: 'it' }],
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
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      output: 'describe("suite", () => { test("foo") })',
      options: [{ fn: 'it', withinDescribe: 'test' }],
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
    },
    {
      code: 'describe("suite", () => { it.only("foo") })',
      output: 'describe("suite", () => { test.only("foo") })',
      options: [{ fn: 'it', withinDescribe: 'test' }],
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
    },
    {
      code: 'describe("suite", () => { xit("foo") })',
      output: 'describe("suite", () => { xtest("foo") })',
      options: [{ fn: 'it', withinDescribe: 'test' }],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'describe("suite", () => { it.skip("foo") })',
      output: 'describe("suite", () => { test.skip("foo") })',
      options: [{ fn: 'it', withinDescribe: 'test' }],
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
    },
    {
      code: 'describe("suite", () => { it.concurrent("foo") })',
      output: 'describe("suite", () => { test.concurrent("foo") })',
      options: [{ fn: 'it', withinDescribe: 'test' }],
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
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      output: 'describe("suite", () => { test("foo") })',
      options: [{ fn: 'test', withinDescribe: 'test' }],
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
    },
    {
      code: 'it("foo")',
      output: 'test("foo")',
      options: [{ fn: 'test', withinDescribe: 'test' }],
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
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      output: 'describe("suite", () => { it("foo") })',
      options: [{ fn: 'it', withinDescribe: 'it' }],
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
    },
    {
      code: 'test("foo")',
      output: 'it("foo")',
      options: [{ fn: 'it', withinDescribe: 'it' }],
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
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      output: 'describe("suite", () => { it("foo") })',
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
    },
    {
      code: 'it("foo")',
      output: 'test("foo")',
      options: [{ withinDescribe: 'it' }],
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
    },
    {
      code: 'describe("suite", () => { test("foo") })',
      output: 'describe("suite", () => { it("foo") })',
      options: [{ withinDescribe: 'it' }],
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
    },
    {
      code: 'it("foo")',
      output: 'test("foo")',
      options: [{ withinDescribe: 'test' }],
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
    },
    {
      code: 'describe("suite", () => { it("foo") })',
      output: 'describe("suite", () => { test("foo") })',
      options: [{ withinDescribe: 'test' }],
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
    },
    {
      code: "test('foo'); // valid\ntest.only('foo'); // valid\n\nit('foo'); // invalid\nit.only('foo'); // invalid",
      output:
        "test('foo'); // valid\ntest.only('foo'); // valid\n\ntest('foo'); // invalid\ntest.only('foo'); // invalid",
      options: [{ fn: 'test' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 3,
        },
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 8,
        },
      ],
    },
    {
      code: "it('foo'); // valid\nit.only('foo'); // valid\n\ntest('foo'); // invalid\ntest.only('foo'); // invalid",
      output:
        "it('foo'); // valid\nit.only('foo'); // valid\n\nit('foo'); // invalid\nit.only('foo'); // invalid",
      options: [{ fn: 'it' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 5,
        },
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 10,
        },
      ],
    },
    {
      code: "it('foo'); // valid\ndescribe('foo', function () {\n  test('bar'); // valid\n});\n\ntest('foo'); // invalid\ndescribe('foo', function () {\n  it('bar'); // invalid\n});",
      output:
        "it('foo'); // valid\ndescribe('foo', function () {\n  test('bar'); // valid\n});\n\nit('foo'); // invalid\ndescribe('foo', function () {\n  test('bar'); // invalid\n});",
      options: [{ fn: 'it', withinDescribe: 'test' }],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 5,
        },
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
        },
      ],
    },
    {
      code: "test('foo'); // valid\ndescribe('foo', function () {\n  it('bar'); // valid\n});\n\nit('foo'); // invalid\ndescribe('foo', function () {\n  test('bar'); // invalid\n});",
      output:
        "test('foo'); // valid\ndescribe('foo', function () {\n  it('bar'); // valid\n});\n\ntest('foo'); // invalid\ndescribe('foo', function () {\n  it('bar'); // invalid\n});",
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 3,
        },
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 7,
        },
      ],
    },
  ],
});
