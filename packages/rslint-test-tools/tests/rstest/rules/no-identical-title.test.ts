import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-identical-title', {} as never, {
  valid: [
    {
      code: 'test("one", () => {}); test("two", () => {});',
    },
    {
      code: 'describe("same", () => {}); test("same", () => {});',
    },
    {
      code: `
        describe("outer", () => {
          test("same", () => {});
          describe("inner", () => {
            test("same", () => {});
          });
        });
      `,
    },
    {
      code: `
        test.each([1, 2])("case %i", () => {});
        test.each([3, 4])("case %i", () => {});
      `,
    },
    {
      code: `
        import { test } from "@rstest/playwright";
        test.describe("a", () => { test("same", () => {}); });
        test.describe("b", () => { test("same", () => {}); });
      `,
    },
  ],
  invalid: [
    {
      code: `
        import.meta.rstest.test("same", () => {});
        import.meta.rstest.test("same", () => {});
      `,
      errors: [{ messageId: 'multipleTestTitle', line: 3, column: 33 }],
    },
    {
      code: `
        import { test } from "@rstest/playwright";
        test.describe("same", () => {});
        test.describe("same", () => {});
      `,
      errors: [{ messageId: 'multipleDescribeTitle', line: 4, column: 23 }],
    },
    {
      code: `
        test("same", () => {});
        test("same", () => {});
      `,
      errors: [{ messageId: 'multipleTestTitle', line: 3, column: 14 }],
    },
    {
      code: `
        describe("same", () => {});
        describe.only("same", () => {});
      `,
      errors: [{ messageId: 'multipleDescribeTitle', line: 3, column: 23 }],
    },
    {
      code: `
        describe("suite", () => {
          test("same", () => {});
          test.only("same", () => {});
        });
      `,
      errors: [{ messageId: 'multipleTestTitle', line: 4, column: 21 }],
    },
  ],
});
