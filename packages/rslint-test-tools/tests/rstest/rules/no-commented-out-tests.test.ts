import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-commented-out-tests', {
  valid: [
    // Active Rstest calls are not comments.
    { code: 'test("foo", () => {})' },
    { code: 'it.skip("foo", () => {})' },
    { code: 'describe.only("foo", () => {})' },
    { code: 'it.only.fails("foo", () => {})' },
    { code: 'describe.only.concurrent("foo", () => {})' },
    {
      code: 'test.each`\nvalue\n${1}\n`("$value", ({ value }) => {})',
    },
    { code: 'describe.each<Row>(rows)("foo", ({ value }) => {})' },

    // Jest aliases are not part of the Rstest API.
    { code: '// fit("foo", () => {})' },
    { code: '// xit("foo", () => {})' },
    { code: '// xtest("foo", () => {})' },
    { code: '// fdescribe("foo", () => {})' },
    { code: '// xdescribe("foo", () => {})' },

    // Similar names and APIs outside the supported Rstest roots are ignored.
    { code: '// suite("foo", () => {})' },
    { code: '// myTest("foo", () => {})' },
    { code: '// rstest.test("foo", () => {})' },
    { code: '// beforeEach(() => {})' },
    { code: '// testSomething()' },
    { code: '// latest(items)' },
    { code: '// test`not a parameterized Rstest API`' },
    { code: '// test.only' },
    { code: '// describe.concurrent' },

    // Type arguments are only recognized after each/for, so prose
    // containing angle brackets is not mistaken for a test call.
    { code: '// test <input> (see docs)' },
    { code: '// it <-> (x)' },
    { code: '// describe <T> (generic explanation)' },

    // Factory calls do not register tests until their returned API is invoked.
    { code: '// test.runIf(condition)' },
    { code: '// test.skipIf(condition).only' },
    { code: '// test.each(rows)' },
    { code: '// test.for`value | expected`' },
    { code: '// describe.each<Row>(rows)' },
    { code: '// test.extend(fixtures)' },
    { code: '// test.extend(fixtures).only' },

    // Unsupported or non-static call chains are ignored.
    { code: '// test.each(rows).only("foo", () => {})' },
    { code: '// describe.for(rows).concurrent("foo", () => {})' },
    { code: '// test.only.extend(fixtures)("foo", () => {})' },
    { code: '// describe.extend(fixtures)("foo", () => {})' },
    { code: '// describe.fails("foo", () => {})' },
    { code: '// test.unknown("foo", () => {})' },
    { code: '// test[modifier]("foo", () => {})' },

    // Documentation and prose are not disabled tests.
    { code: '/** test("documented example", () => {}) */' },
    {
      code: '// ```ts\n// test("documented example", () => {})\n// ```',
    },
    { code: '// test("foo") should be preferred in examples' },
  ],
  invalid: [
    // Direct test and suite calls.
    {
      code: '// test("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '// it("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '// describe("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },

    // Every Rstest test modifier.
    {
      code: '// test.only("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '// test.skip("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '// test.todo("foo")',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '// test.fails("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '// test.concurrent("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '// test.sequential("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '// test.runIf(condition)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '// test.skipIf(condition)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },

    // Every Rstest suite modifier.
    {
      code: '// describe.only("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '// describe.skip("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '// describe.todo("foo")',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '// describe.concurrent("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '// describe.sequential("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '// describe.runIf(condition)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: '// describe.skipIf(condition)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },

    // Getter and conditional modifiers can be chained.
    {
      code: '// it.only.fails("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '// test.skip.concurrent("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '// test.concurrent.only("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '// test.concurrent.runIf(condition)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '// describe.only.concurrent("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: '// describe.concurrent.only("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: '// describe.concurrent.skipIf(condition)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    {
      code: '// describe.skipIf(condition).concurrent("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },

    // Array-based parameterized tests and suites.
    {
      code: '// test.each(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: '// test.for(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: '// it.each(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: '// it.for(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '// describe.each(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: '// describe.for(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '// it.only.each(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '// it.concurrent.for(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '// describe.only.each(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: '// describe.skip.for(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },

    // Explicit type arguments are supported by each and for.
    {
      code: '// test.each<Row>(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '// test.for<Row>(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: '// describe.each<Row>(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: '// describe.for<Row>(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '// test.for<{ value: Map<string, number> }>(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 76,
        },
      ],
    },

    // Tagged-template parameterized tests and suites.
    {
      code: '// test.each`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '// test.for`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: '// it.each`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '// it.for`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: '// describe.each`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: '// describe.for`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: '// test.for<Row>`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: '// describe.each<Row>`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 66,
        },
      ],
    },
    {
      code: '// it.concurrent.for<Row>`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },

    // Property and bracket access can be chained.
    {
      code: '// describe["skip"]("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '// test["only"]["concurrent"]("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: '// describe["only"].each(rows)("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: '// test["for"]`value | expected`("foo", ({ value }) => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },

    // Block comments may contain multiline calls and chains.
    {
      code: '/*\n  describe("foo", () => {})\n*/',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 3,
        },
      ],
    },
    {
      code: '/*\n  describe\n    .only\n    .concurrent("foo", () => {})\n*/',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 3,
        },
      ],
    },
    {
      code: '/*\n  test.for<Row>`\n    value\n    ${1}\n  `("$value", ({ value }) => {})\n*/',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 6,
          endColumn: 3,
        },
      ],
    },

    // Conditional and extended APIs must eventually register a test.
    {
      code: '// test.runIf(condition).only("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: '// describe.only.skipIf(condition)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '// test.extend(fixtures)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '// it.extend<Fixtures>(fixtures).only("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '// test.extend(a).extend(b)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: '// test.extend(fixtures).each(rows)("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },

    // Transparent syntax and static property access.
    {
      code: '// ((test))("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '// test?.("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '// test?.only("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '// test[`only`]("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },

    // Consecutive line comments are parsed as one logical code block.
    {
      code: '// test\n//   .only\n//   .concurrent("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '// test.each<Row>`\n// value | expected\n// ${1} | ${2}\n// `("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '// test.extend({\n//   value: 1,\n// }).only("foo", () => {})',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '/* test("renders", () => <div />) */',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '// test("first", () => {}); cleanup()',
      errors: [
        {
          messageId: 'commentedTests',
          message: 'Do not comment out tests',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
  ],
});
