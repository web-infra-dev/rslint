import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-disabled-tests', {} as never, {
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
  ],
  invalid: [
    {
      code: 'test.skip("case", () => {});',
      errors: [
        {
          line: 1,
          column: 1,
          messageId: 'skippedTest',
          message: 'Tests should not be skipped',
        },
      ],
    },
    {
      code: 'it.skip("case", () => {});',
      errors: [{ line: 1, column: 1, messageId: 'skippedTest' }],
    },
    {
      code: 'describe.skip("suite", () => {});',
      errors: [{ line: 1, column: 1, messageId: 'skippedTest' }],
    },
    {
      code: 'test["skip"]("case", () => {}); it[`skip`]("case", () => {});',
      errors: [
        { line: 1, column: 1, messageId: 'skippedTest' },
        { line: 1, column: 33, messageId: 'skippedTest' },
      ],
    },
    {
      code: 'test.concurrent.skip("case", () => {}); test.skip.sequential("case", () => {});',
      errors: [
        { line: 1, column: 1, messageId: 'skippedTest' },
        { line: 1, column: 41, messageId: 'skippedTest' },
      ],
    },
    {
      code: 'test.skip.each([1])("case", () => {}); test.skip.for([1])("case", () => {});',
      errors: [
        { line: 1, column: 1, messageId: 'skippedTest' },
        { line: 1, column: 40, messageId: 'skippedTest' },
      ],
    },
    {
      code: 'describe.skip.each([1])("suite", () => {});',
      errors: [{ line: 1, column: 1, messageId: 'skippedTest' }],
    },
    {
      code: [
        'import { test as rstestTest } from "@rstest/core";',
        'rstestTest.skip("case", () => {});',
      ].join('\n'),
      errors: [{ line: 2, column: 1, messageId: 'skippedTest' }],
    },
    {
      code: [
        'import * as rstest from "@rstest/core";',
        'rstest.test.skip("case", () => {});',
        'rstest.describe.skip("suite", () => {});',
      ].join('\n'),
      errors: [
        { line: 2, column: 1, messageId: 'skippedTest' },
        { line: 3, column: 1, messageId: 'skippedTest' },
      ],
    },
    {
      code: ['const skipped = test.skip;', 'skipped("case", () => {});'].join(
        '\n',
      ),
      errors: [{ line: 2, column: 1, messageId: 'skippedTest' }],
    },
    {
      code: 'test.extend({}).skip("case", () => {});',
      errors: [{ line: 1, column: 1, messageId: 'skippedTest' }],
    },
    {
      code: 'test("case"); it("case"); test.concurrent("case");',
      errors: [
        {
          line: 1,
          column: 1,
          messageId: 'missingFunction',
          message: 'Test is missing function argument',
        },
        { line: 1, column: 15, messageId: 'missingFunction' },
        { line: 1, column: 27, messageId: 'missingFunction' },
      ],
    },
    {
      code: 'test.each([1])("case"); test.for([1])("case");',
      errors: [
        { line: 1, column: 1, messageId: 'missingFunction' },
        { line: 1, column: 25, messageId: 'missingFunction' },
      ],
    },
    // `(description, options)` registers a test without a body
    {
      code: 'test("case", { timeout: 100 }); it("case", {});',
      errors: [
        { line: 1, column: 1, messageId: 'missingFunction' },
        { line: 1, column: 33, messageId: 'missingFunction' },
      ],
    },
    // both a skipped test and a missing implementation
    {
      code: 'test.skip("case");',
      errors: [
        { line: 1, column: 1, messageId: 'skippedTest' },
        { line: 1, column: 1, messageId: 'missingFunction' },
      ],
    },
    // `.todo` suppresses only the missing-function report
    {
      code: 'test.todo.skip("case");',
      errors: [{ line: 1, column: 1, messageId: 'skippedTest' }],
    },
  ],
});
