import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-commented-out-tests', {} as never, {
  valid: [
    { code: 'test("foo", () => {})' },
    { code: 'it.skip("foo", () => {})' },
    { code: 'describe.only("foo", () => {})' },
    { code: '// fit("foo", () => {})' },
    { code: '// xit("foo", () => {})' },
    { code: '// xtest("foo", () => {})' },
    { code: '// fdescribe("foo", () => {})' },
    { code: '// xdescribe("foo", () => {})' },
    { code: '// testSomething()' },
    { code: '// latest(items)' },
  ],
  invalid: [
    { code: '// test("foo", () => {})', errors: 1 },
    { code: '// it.skip("foo", () => {})', errors: 1 },
    { code: '// describe.only("foo", () => {})', errors: 1 },
    {
      code: '// test.for([{ value: 1 }])("$value", ({ value }) => {})',
      errors: 1,
    },
    { code: '// describe["skip"]("foo", () => {})', errors: 1 },
    {
      code: '/*\n  describe("foo", () => {})\n*/',
      errors: [{ message: 'Do not comment out tests' }],
    },
  ],
});
