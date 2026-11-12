import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-disabled-tests', {
  valid: [
    {
      code: 'test("case", () => {}); it("case", () => {}); describe("suite", () => {});',
    },
    { code: 'test.only("case", () => {}); test.fails("case", () => {});' },
    {
      code: 'describe.only("suite", () => {}); describe.concurrent.sequential("suite", () => {});',
    },
    { code: 'test.todo("case"); it.todo("case"); describe.todo("suite");' },
    { code: 'test.todo.each([1])("case"); test.todo.for([1])("case");' },
    { code: 'const todoTest = test.todo; todoTest("case");' },
    {
      code: 'const condition = true; test.skipIf(condition)("case", () => {}); test.runIf(condition)("case", () => {});',
    },
    {
      code: 'const condition = true; describe.skipIf(condition)("suite", () => {}); describe.runIf(condition)("suite", () => {});',
    },
    // `skip` on the test context is a runtime decision, not a disabled test.
    { code: 'test("case", context => { context.skip(); });' },
    { code: 'describe("empty suite"); describe.concurrent("empty suite");' },
    { code: 'test("case", { timeout: 100 }, () => {});' },
    // an identifier is not necessarily an options object
    { code: 'declare const options: any; test("case", options);' },
    // jest-only aliases are not rstest APIs
    {
      code: 'declare const fit: any, xit: any, xtest: any, fdescribe: any, xdescribe: any; fit("case", () => {}); xit("case", () => {}); xtest("case", () => {}); fdescribe("suite", () => {}); xdescribe("suite", () => {});',
    },
    { code: 'declare const pending: any; pending();' },
    // dynamic member access cannot be resolved to `skip`
    {
      code: 'declare const suffix: string; test[`sk${suffix}`]("case", () => {});',
    },
    // not the rstest `test`
    { code: 'import { test } from "node:test"; test.skip("case", () => {});' },
    {
      code: 'import { test, describe } from "vitest"; test.skip("case", () => {}); describe.skip("suite", () => {});',
    },
    {
      code: 'declare function createRunner(): any; const test = createRunner(); test.skip("case", () => {});',
    },
    // `.skip` after `.each`/`.for` is not a valid rstest chain
    {
      code: 'test.for([1]).skip("case", () => {}); describe.each([1]).skip("suite", () => {});',
    },
    { code: 'import.meta.rstest.test.todo("case");' },
    {
      code: 'import { test } from "@rstest/playwright"; test("case", () => {});',
    },
    {
      code: 'import { it } from "@rstest/playwright"; it.skip("case", () => {});',
    },
  ],
  invalid: [
    {
      code: 'import.meta.rstest.test.skip("case", () => {});',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: [
        'import { test } from "@rstest/playwright";',
        'test.skip("case", () => {});',
      ].join('\n'),
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'test.skip("case", () => {});',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'it.skip("case", () => {});',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'describe.skip("suite", () => {});',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'test["skip"]("case", () => {}); it[`skip`]("case", () => {});',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: 'test.concurrent.skip("case", () => {}); test.skip.sequential("case", () => {});',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 79,
        },
      ],
    },
    {
      code: 'test.skip.each([1])("case", () => {}); test.skip.for([1])("case", () => {});',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 76,
        },
      ],
    },
    {
      code: 'describe.skip.each([1])("suite", () => {});',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: [
        'import { test as rstestTest } from "@rstest/core";',
        'rstestTest.skip("case", () => {});',
      ].join('\n'),
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
    {
      code: [
        'import * as rstest from "@rstest/core";',
        'rstest.test.skip("case", () => {});',
        'rstest.describe.skip("suite", () => {});',
      ].join('\n'),
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 40,
        },
      ],
    },
    {
      code: ['const skipped = test.skip;', 'skipped("case", () => {});'].join(
        '\n',
      ),
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'test.extend({}).skip("case", () => {});',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'test("case"); it("case"); test.concurrent("case");',
      errors: [
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'test.each([1])("case"); test.for([1])("case");',
      errors: [
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    // `(description, options)` registers a test without a body
    {
      code: 'test("case", { timeout: 100 }); it("case", {});',
      errors: [
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    // both a skipped test and a missing implementation
    {
      code: 'test.skip("case");',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    // `.todo` suppresses only the missing-function report
    {
      code: 'test.todo.skip("case");',
      errors: [
        {
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
  ],
});
