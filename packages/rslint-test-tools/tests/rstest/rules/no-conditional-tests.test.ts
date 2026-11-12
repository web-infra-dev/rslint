import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-conditional-tests', {
  valid: [
    { code: 'test("shows error", () => {});' },
    { code: 'it("foo", function () {})' },
    { code: "it('foo', () => {}); function myTest() { if ('bar') {} }" },
    {
      code: `function myFunc(str: string) {
    return str;
  }
  describe("myTest", () => {
     it("convert shortened equal filter", () => {
      expect(
      myFunc("5")
      ).toEqual("5");
     });
    });`,
    },
    {
      code: `describe("shows error", () => {
     if (1 === 2) {
      myFunc();
     }
     expect(true).toBe(false);
    });`,
    },
  ],
  invalid: [
    {
      code: `describe("shows error", () => {
    if(true) {
     test("shows error", () => {
      expect(true).toBe(true);
     })
    }
   })`,
      errors: [
        {
          messageId: 'noConditionalTests',
          message: 'Avoid using if conditions in a test',
          line: 3,
          column: 6,
          endLine: 3,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
   describe("shows error", () => {
    if(true) {
     it("shows error", () => {
      expect(true).toBe(true);
      })
     }
   })`,
      errors: [
        {
          messageId: 'noConditionalTests',
          message: 'Avoid using if conditions in a test',
          line: 4,
          column: 6,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    {
      code: `describe("errors", () => {
    if (Math.random() > 0) {
     test("test2", () => {
     expect(true).toBeTruthy();
    });
    }
   });`,
      errors: [
        {
          messageId: 'noConditionalTests',
          message: 'Avoid using if conditions in a test',
          line: 3,
          column: 6,
          endLine: 3,
          endColumn: 10,
        },
      ],
    },
  ],
});
