import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-test-timeout', {} as never, {
  valid: [
    { code: `test.todo("a")` },
    { code: `xit("a", () => {})` },
    { code: `test("a", () => {}, 0)` },
    { code: `it("a", () => {}, 500)` },
    { code: `it.skip("a", () => {})` },
    { code: `test.skip("a", () => {})` },
    { code: `test("a", () => {}, 1000)` },
    { code: `it.only("a", () => {}, 1234)` },
    { code: `test.only("a", () => {}, 1234)` },
    { code: `it.concurrent("a", () => {}, 400)` },
    { code: `test("a", () => {}, { timeout: 0 })` },
    { code: `test.concurrent("a", () => {}, 400)` },
    { code: `test("a", () => {}, { timeout: 500 })` },
    { code: `test("a", { timeout: 500 }, () => {})` },
    { code: `const t = 500; test("a", { timeout: t }, () => {})` },
    { code: `const t = 500; test("a", () => {}, t)` },
    { code: `const opts = { timeout: 500 }; test("a", opts, () => {})` },
    {
      code: `const T = 1000; rs.setConfig({ testTimeout: T }); test("a", () => {})`,
    },
    { code: `rs.setConfig({ testTimeout: 1000 }); test("a", () => {})` },
    { code: `const t = 500; it("a", { timeout: t }, () => {})` },
    { code: `const t = 500; it("a", () => {}, t)` },
    { code: `const opts = { timeout: 500 }; it("a", opts, () => {})` },
    {
      code: `const T = 1000; rs.setConfig({ testTimeout: T }); it("a", () => {})`,
    },
    { code: `rs.setConfig({ testTimeout: 1000 }); it("a", () => {})` },
    { code: `test("a", { foo: 1 }, { timeout: 500 }, () => {})` },
    { code: `test("a", { timeout: 500 }, 1000, () => {})` },
    { code: `test("a", () => {}, 1000, { extra: true })` },
    {
      code: `if (import.meta.rstest) { const opts = { timeout: 500 }; describe("outer", () => { it("repro: same-file opts object", opts, () => {}); }); }`,
    },
    {
      code: `if (import.meta.rstest) { const T = 500; describe("outer", () => { describe("inner", () => { it("repro: same-file const timeout", () => {}, T); }); }); }`,
    },
    {
      code: `import { TIMEOUT } from "./test-constants"; test("a", () => {}, TIMEOUT)`,
    },
    {
      code: `import { TIMEOUT } from "./test-constants"; it("a", () => {}, TIMEOUT)`,
    },
    {
      code: `import { TIMEOUT } from "./test-constants"; test("a", () => {}, { timeout: TIMEOUT })`,
    },
    {
      code: `import { TIMEOUT } from "./test-constants"; test("a", { timeout: TIMEOUT }, () => {})`,
    },
    {
      code: `import { OPTS } from "./test-constants"; test("a", OPTS, () => {})`,
    },
    { code: `import T from "./test-constants"; test("a", () => {}, T)` },
    { code: `let t = 500; test("a", () => {}, t)` },
    { code: `let t = 500; it("a", () => {}, t)` },
    { code: `const t = getTimeout(); test("a", () => {}, t)` },
    { code: `const t = getTimeout(); it("a", () => {}, t)` },
    { code: `test("a", () => {}, { timeout: null })` },
    { code: `test("a", () => {}, { timeout: undefined })` },
    { code: `rs.setConfig({ testTimeout: null }); test("a", () => {})` },
    { code: `rs.setConfig({ testTimeout: undefined }); test("a", () => {})` },
    {
      code: `const opts = { timeout: 1000 }; test("a", { ...opts }, () => {})`,
    },
    {
      code: `const opts = { timeout: 1000 }; test("a", { ...opts }, { foo: 1 }, () => {})`,
    },
    { code: `test("a", () => {}, { timeout: -1 }, { timeout: 500 })` },
    { code: `test("a", () => {}, { timeout: 500 }, { timeout: -1 })` },
    { code: `test("a", () => {}, { timeout: -1 }, 1000)` },
    { code: `test("a", () => {}, 1000, { timeout: -1 })` },
  ],
  invalid: [
    {
      code: `test("a", () => {})`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: `it("a", () => {})`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: `test.only("a", () => {})`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: `test.concurrent("a", () => {})`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: `it.concurrent("a", () => {})`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: `rs.setConfig({}); test("a", () => {})`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: `test("a", () => {}, -100)`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: `test("a", () => {}, { timeout: -1 })`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: `test("a", () => {}); rs.setConfig({ testTimeout: 1000 })`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: `import { TIMEOUT } from "./test-constants"; test("a", () => {})`,
      errors: [
        {
          messageId: 'missingTimeout',
          message: 'Test is missing a timeout. Add an explicit timeout.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 64,
        },
      ],
    },
  ],
});
