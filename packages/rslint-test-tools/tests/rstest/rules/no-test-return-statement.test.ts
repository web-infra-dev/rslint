import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-test-return-statement', {} as never, {
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
      errors: [{ messageId: 'noTestReturnStatement', column: 7, line: 2 }],
    },
    {
      code: `it("one", function () {
      return expect(1).toBe(1);
       });`,
      errors: [{ messageId: 'noTestReturnStatement', column: 7, line: 2 }],
    },
    {
      code: `it.skip("one", function () {
      return expect(1).toBe(1);
       });`,
      errors: [{ messageId: 'noTestReturnStatement', column: 7, line: 2 }],
    },
    {
      code: `it("one", myTest);
     function myTest () {
       return expect(1).toBe(1);
     }`,
      errors: [{ messageId: 'noTestReturnStatement', column: 8, line: 3 }],
    },
    {
      code: `
        import { test } from '@rstest/core';

        test('options', { timeout: 100 }, () => {
          return 1;
        });
      `,
      errors: [{ messageId: 'noTestReturnStatement', column: 11, line: 5 }],
    },
  ],
});
