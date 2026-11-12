import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-todo', {
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
      code: 'const failing = test.fails.skip; failing("stub", () => {});',
    },
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
      code: 'test?.("optional");',
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'test("extra", () => {}, 1, 2);',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'test("invalid timeout", () => {}, "not a timeout");',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'test("invalid timeout", () => {}, {});',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'test("invalid timeout", () => {}, null);',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'test("invalid timeout", () => {}, ("bad" as any));',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'test(`dynamic ${name}`);',
      output: 'test.todo(`dynamic ${name}`);',
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'wrap(test("nested"));',
      output: 'wrap(test.todo("nested"));',
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      // Reported without a fix: rewriting either accessor leaves the other
      // skip active, so no one-accessor rewrite delivers a todo.
      code: 'test.skip.skip("stub", () => {});',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'const skipped = test.skip; skipped.skip("stub", () => {});',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    {
      code: [
        'const { test } = import.meta.rstest;',
        'test("i need to write this test");',
      ].join('\n'),
      output: [
        'const { test } = import.meta.rstest;',
        'test.todo("i need to write this test");',
      ].join('\n'),
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      code: `test("i need to write this test");`,
      output: 'test.todo("i need to write this test");',
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: `test("i need to write this test",);`,
      output: 'test.todo("i need to write this test",);',
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'test(`i need to write this test`);',
      output: 'test.todo(`i need to write this test`);',
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'it("foo", function () {})',
      output: 'it.todo("foo")',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'it("foo", () => {})',
      output: 'it.todo("foo")',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: `test.skip("i need to write this test", () => {});`,
      output: 'test.todo("i need to write this test");',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: `test.skip("i need to write this test", function() {});`,
      output: 'test.todo("i need to write this test");',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: `test["skip"]("i need to write this test", function() {});`,
      output: 'test["todo"]("i need to write this test");',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: 'test[`skip`]("i need to write this test", function() {});',
      output: 'test[`todo`]("i need to write this test");',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: 'test("x", { timeout: 1_000 });',
      output: 'test.todo("x", { timeout: 1_000 });',
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'test("x", () => {}, 1_000);',
      output: 'test.todo("x");',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'test("x", { retry: 2 }, () => {});',
      output: 'test.todo("x", { retry: 2 });',
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
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
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 17,
        },
      ],
    },
    {
      code: ['const api = import.meta.rstest;', 'api.test("x");'].join('\n'),
      output: ['const api = import.meta.rstest;', 'api.test.todo("x");'].join(
        '\n',
      ),
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
  ],
});
