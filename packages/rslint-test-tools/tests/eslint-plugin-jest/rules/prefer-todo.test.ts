import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-todo', {} as never, {
  valid: [
    { code: 'test()' },
    { code: 'test.concurrent()' },
    { code: 'test.todo("i need to write this test");' },
    { code: 'test(obj)' },
    { code: 'test.concurrent(obj)' },
    { code: 'fit("foo")' },
    { code: 'fit.concurrent("foo")' },
    { code: 'xit("foo")' },
    { code: 'test("foo", 1)' },
    { code: 'test("stub", () => expect(1).toBe(1));' },
    { code: 'test.concurrent("stub", () => expect(1).toBe(1));' },
    {
      code: `
        supportsDone && params.length < test.length
          ? done => test(...params, done)
          : () => test(...params);
      `,
    },
  ],
  invalid: [
    {
      code: `test("i need to write this test");`,
      output: 'test.todo("i need to write this test");',
      errors: [{ messageId: 'unimplementedTest' }],
    },
    {
      code: `test("i need to write this test",);`,
      output: 'test.todo("i need to write this test",);',
      errors: [{ messageId: 'unimplementedTest' }],
    },
    {
      code: 'test(`i need to write this test`);',
      output: 'test.todo(`i need to write this test`);',
      errors: [{ messageId: 'unimplementedTest' }],
    },
    {
      code: 'it("foo", function () {})',
      output: 'it.todo("foo")',
      errors: [{ messageId: 'emptyTest' }],
    },
    {
      code: 'it("foo", () => {})',
      output: 'it.todo("foo")',
      errors: [{ messageId: 'emptyTest' }],
    },
    {
      code: `test.skip("i need to write this test", () => {});`,
      output: 'test.todo("i need to write this test");',
      errors: [{ messageId: 'emptyTest' }],
    },
    {
      code: `test.skip("i need to write this test", function() {});`,
      output: 'test.todo("i need to write this test");',
      errors: [{ messageId: 'emptyTest' }],
    },
    {
      code: `test["skip"]("i need to write this test", function() {});`,
      output: 'test[\'todo\']("i need to write this test");',
      errors: [{ messageId: 'emptyTest' }],
    },
    {
      code: `test[\`skip\`]("i need to write this test", function() {});`,
      output: 'test[\'todo\']("i need to write this test");',
      errors: [{ messageId: 'emptyTest' }],
    },
  ],
});
