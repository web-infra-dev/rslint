import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-test-return-statement', {
  valid: [
    { code: 'it("noop", function () {});' },
    { code: 'test("noop", () => {});' },
    { code: 'test("one", () => expect(1).toBe(1));' },
    { code: 'test("empty")' },
    {
      code: `it("one", myTest);
    function myTest() {
      expect(1).toBe(1);
    }`,
    },
    {
      code: `it("one", () => expect(1).toBe(1));
       function myHelper() {}`,
    },
  ],
  invalid: [
    {
      code: `test("one", () => {
      return expect(1).toBe(1);
       });`,
      errors: [
        {
          messageId: 'noTestReturnStatement',
          message: 'Return statements are not allowed in tests',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 32,
        },
      ],
    },
    {
      code: `it("one", function () {
      return expect(1).toBe(1);
       });`,
      errors: [
        {
          messageId: 'noTestReturnStatement',
          message: 'Return statements are not allowed in tests',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 32,
        },
      ],
    },
    {
      code: `it.skip("one", function () {
      return expect(1).toBe(1);
       });`,
      errors: [
        {
          messageId: 'noTestReturnStatement',
          message: 'Return statements are not allowed in tests',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 32,
        },
      ],
    },
    {
      code: `it("one", myTest);
     function myTest () {
       return expect(1).toBe(1);
     }`,
      errors: [
        {
          messageId: 'noTestReturnStatement',
          message: 'Return statements are not allowed in tests',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 33,
        },
      ],
    },
    {
      code: `
        import { test } from '@rstest/core';

        test('options', { timeout: 100 }, () => {
          return 1;
        });
      `,
      errors: [
        {
          messageId: 'noTestReturnStatement',
          message: 'Return statements are not allowed in tests',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 20,
        },
      ],
    },
  ],
});
