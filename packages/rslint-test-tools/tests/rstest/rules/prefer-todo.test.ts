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
    { code: 'test.skip.skip("stub", () => {});' },
    {
      code: [
        'const core = require("@rstest/core");',
        'core.test = custom;',
        'core.test("stub");',
      ].join('\n'),
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
      output: 'test["todo"]("i need to write this test");',
      errors: [{ messageId: 'emptyTest' }],
    },
    {
      code: 'test[`skip`]("i need to write this test", function() {});',
      output: 'test[`todo`]("i need to write this test");',
      errors: [{ messageId: 'emptyTest' }],
    },
    {
      code: 'test("x", { timeout: 1_000 });',
      output: 'test.todo("x", { timeout: 1_000 });',
      errors: [{ messageId: 'unimplementedTest' }],
    },
    {
      code: 'test("x", () => {}, 1_000);',
      output: 'test.todo("x");',
      errors: [{ messageId: 'emptyTest' }],
    },
    {
      code: 'test("x", { retry: 2 }, () => {});',
      output: 'test.todo("x", { retry: 2 });',
      errors: [{ messageId: 'emptyTest' }],
    },
    {
      code: [
        'import * as rstest from "@rstest/core";',
        'rstest.test("x");',
      ].join('\n'),
      output: [
        'import * as rstest from "@rstest/core";',
        'rstest.test.todo("x");',
      ].join('\n'),
      errors: [{ messageId: 'unimplementedTest' }],
    },
    {
      code: ['const api = import.meta.rstest;', 'api.test("x");'].join('\n'),
      output: ['const api = import.meta.rstest;', 'api.test.todo("x");'].join(
        '\n',
      ),
      errors: [{ messageId: 'unimplementedTest' }],
    },
  ],
});
