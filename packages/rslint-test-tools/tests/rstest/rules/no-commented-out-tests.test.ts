import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const invalid = (code: string) => ({
  code,
  errors: [{ message: 'Do not comment out tests' }],
});

ruleTester.run('no-commented-out-tests', {} as never, {
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
  ],
  invalid: [
    // Direct test and suite calls.
    invalid('// test("foo", () => {})'),
    invalid('// it("foo", () => {})'),
    invalid('// describe("foo", () => {})'),

    // Every Rstest test modifier.
    invalid('// test.only("foo", () => {})'),
    invalid('// test.skip("foo", () => {})'),
    invalid('// test.todo("foo")'),
    invalid('// test.fails("foo", () => {})'),
    invalid('// test.concurrent("foo", () => {})'),
    invalid('// test.sequential("foo", () => {})'),
    invalid('// test.runIf(condition)("foo", () => {})'),
    invalid('// test.skipIf(condition)("foo", () => {})'),

    // Every Rstest suite modifier.
    invalid('// describe.only("foo", () => {})'),
    invalid('// describe.skip("foo", () => {})'),
    invalid('// describe.todo("foo")'),
    invalid('// describe.concurrent("foo", () => {})'),
    invalid('// describe.sequential("foo", () => {})'),
    invalid('// describe.runIf(condition)("foo", () => {})'),
    invalid('// describe.skipIf(condition)("foo", () => {})'),

    // Getter and conditional modifiers can be chained.
    invalid('// it.only.fails("foo", () => {})'),
    invalid('// test.skip.concurrent("foo", () => {})'),
    invalid('// test.concurrent.only("foo", () => {})'),
    invalid('// test.concurrent.runIf(condition)("foo", () => {})'),
    invalid('// describe.only.concurrent("foo", () => {})'),
    invalid('// describe.concurrent.only("foo", () => {})'),
    invalid('// describe.concurrent.skipIf(condition)("foo", () => {})'),
    invalid('// describe.skipIf(condition).concurrent("foo", () => {})'),

    // Array-based parameterized tests and suites.
    invalid('// test.each(rows)("foo", ({ value }) => {})'),
    invalid('// test.for(rows)("foo", ({ value }) => {})'),
    invalid('// it.each(rows)("foo", ({ value }) => {})'),
    invalid('// it.for(rows)("foo", ({ value }) => {})'),
    invalid('// describe.each(rows)("foo", ({ value }) => {})'),
    invalid('// describe.for(rows)("foo", ({ value }) => {})'),
    invalid('// it.only.each(rows)("foo", ({ value }) => {})'),
    invalid('// it.concurrent.for(rows)("foo", ({ value }) => {})'),
    invalid('// describe.only.each(rows)("foo", ({ value }) => {})'),
    invalid('// describe.skip.for(rows)("foo", ({ value }) => {})'),

    // Explicit type arguments are supported by each and for.
    invalid('// test.each<Row>(rows)("foo", ({ value }) => {})'),
    invalid('// test.for<Row>(rows)("foo", ({ value }) => {})'),
    invalid('// describe.each<Row>(rows)("foo", ({ value }) => {})'),
    invalid('// describe.for<Row>(rows)("foo", ({ value }) => {})'),
    invalid(
      '// test.for<{ value: Map<string, number> }>(rows)("foo", ({ value }) => {})',
    ),

    // Tagged-template parameterized tests and suites.
    invalid('// test.each`value | expected`("foo", ({ value }) => {})'),
    invalid('// test.for`value | expected`("foo", ({ value }) => {})'),
    invalid('// it.each`value | expected`("foo", ({ value }) => {})'),
    invalid('// it.for`value | expected`("foo", ({ value }) => {})'),
    invalid('// describe.each`value | expected`("foo", ({ value }) => {})'),
    invalid('// describe.for`value | expected`("foo", ({ value }) => {})'),
    invalid('// test.for<Row>`value | expected`("foo", ({ value }) => {})'),
    invalid(
      '// describe.each<Row>`value | expected`("foo", ({ value }) => {})',
    ),
    invalid(
      '// it.concurrent.for<Row>`value | expected`("foo", ({ value }) => {})',
    ),

    // Property and bracket access can be chained.
    invalid('// describe["skip"]("foo", () => {})'),
    invalid('// test["only"]["concurrent"]("foo", () => {})'),
    invalid('// describe["only"].each(rows)("foo", ({ value }) => {})'),
    invalid('// test["for"]`value | expected`("foo", ({ value }) => {})'),

    // Block comments may contain multiline calls and chains.
    invalid('/*\n  describe("foo", () => {})\n*/'),
    invalid('/*\n  describe\n    .only\n    .concurrent("foo", () => {})\n*/'),
    invalid(
      '/*\n  test.for<Row>`\n    value\n    ${1}\n  `("$value", ({ value }) => {})\n*/',
    ),
  ],
});
