import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-conditional-tests', {} as never, {
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
      errors: [{ messageId: 'noConditionalTests', line: 3, column: 6 }],
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
      errors: [{ messageId: 'noConditionalTests', line: 4, column: 6 }],
    },
    {
      code: `describe("errors", () => {
    if (Math.random() > 0) {
     test("test2", () => {
     expect(true).toBeTruthy();
    });
    }
   });`,
      errors: [{ messageId: 'noConditionalTests', line: 3, column: 6 }],
    },
  ],
});
