import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-identical-title', {
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
      errors: [
        {
          messageId: 'multipleTestTitle',
          message:
            'Test title is used multiple times in the same describe block',
          line: 3,
          column: 33,
          endLine: 3,
          endColumn: 39,
        },
      ],
    },
    {
      code: `
        import { test } from "@rstest/playwright";
        test.describe("same", () => {});
        test.describe("same", () => {});
      `,
      errors: [
        {
          messageId: 'multipleDescribeTitle',
          message:
            'Describe block title is used multiple times in the same describe block',
          line: 4,
          column: 23,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
        test("same", () => {});
        test("same", () => {});
      `,
      errors: [
        {
          messageId: 'multipleTestTitle',
          message:
            'Test title is used multiple times in the same describe block',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
        describe("same", () => {});
        describe.only("same", () => {});
      `,
      errors: [
        {
          messageId: 'multipleDescribeTitle',
          message:
            'Describe block title is used multiple times in the same describe block',
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
        describe("suite", () => {
          test("same", () => {});
          test.only("same", () => {});
        });
      `,
      errors: [
        {
          messageId: 'multipleTestTitle',
          message:
            'Test title is used multiple times in the same describe block',
          line: 4,
          column: 21,
          endLine: 4,
          endColumn: 27,
        },
      ],
    },
  ],
});
