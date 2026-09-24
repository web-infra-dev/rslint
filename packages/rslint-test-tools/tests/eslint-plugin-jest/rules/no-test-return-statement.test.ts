import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-test-return-statement', {} as never, {
  valid: [
    'it("noop", function () {});',
    'test("noop", () => {});',
    'test("one", () => expect(1).toBe(1));',
    'test("empty")',
    `test("one", () => {
  expect(1).toBe(1);
});`,
    `it("one", function () {
  expect(1).toBe(1);
});`,
    `it("one", myTest);
function myTest() {
  expect(1).toBe(1);
}`,
    `it("one", () => expect(1).toBe(1));
function myHelper() {}`,
  ],
  invalid: [
    {
      code: `test("one", () => {
  return expect(1).toBe(1);
});`,
      errors: [{ messageId: 'noReturnValue', column: 3, line: 2 }],
    },
    {
      code: `it("one", function () {
  return expect(1).toBe(1);
});`,
      errors: [{ messageId: 'noReturnValue', column: 3, line: 2 }],
    },
    {
      code: `it.skip("one", function () {
  return expect(1).toBe(1);
});`,
      errors: [{ messageId: 'noReturnValue', column: 3, line: 2 }],
    },
    {
      code: `it.each\`\`("one", function () {
  return expect(1).toBe(1);
});`,
      errors: [{ messageId: 'noReturnValue', column: 3, line: 2 }],
    },
    {
      code: `it.each()("one", function () {
  return expect(1).toBe(1);
});`,
      errors: [{ messageId: 'noReturnValue', column: 3, line: 2 }],
    },
    {
      code: `it.only.each\`\`("one", function () {
  return expect(1).toBe(1);
});`,
      errors: [{ messageId: 'noReturnValue', column: 3, line: 2 }],
    },
    {
      code: `it.only.each()("one", function () {
  return expect(1).toBe(1);
});`,
      errors: [{ messageId: 'noReturnValue', column: 3, line: 2 }],
    },
    {
      code: `it("one", myTest);
function myTest () {
  return expect(1).toBe(1);
}`,
      errors: [{ messageId: 'noReturnValue', column: 3, line: 3 }],
    },
  ],
});
