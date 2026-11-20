// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-export', {
  valid: [
    ...(
      [
        'describe("a test", () => { expect(1).toBe(1); })',
        'window.location = "valid"',
        'module.somethingElse = "foo";',
        'export const myThing = "valid"',
        'export default function () {}',
        'module.exports = function(){}',
        'module.exports.myThing = "valid";',
        'module.export.foo = "valid"; test("a test", () => {});',
        'const exports = "exports"; module[exports] = {}; test("a test", () => {});',
        'const module = { exports: {} }; module.exports.foo = "valid"; test("a test", () => {});',
        'const run = (module: { exports: object }) => { module.exports.foo = "valid" }; test("a test", () => {});',
        'const exports = { foo: "" }; exports.foo = "valid"; test("a test", () => {});',
        'const run = (exports: { foo: string }) => { exports.foo = "valid" }; test("a test", () => {});',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'describe("a test", () => { expect(1).toBe(1); })',
        'window.location = "valid"',
        'module.somethingElse = "foo";',
        'my.exports.myThing = "valid";',
        'report.myThing = "valid";',
        'export const myThing = "valid"',
        'export default function () {}',
        'module.exports = function(){}',
        'module["exports"] = function(){}',
        'module.exports.exports.exports = function(){}',
        '["module"].exports.exports.exports = function(){}',
        'module["exports"].exports.exports = function(){}',
        'module["exports"]["exports"].exports = function(){}',
        'module["exports"]["exports"]["exports"] = function(){}',
        '["exports"]["exports"]["exports"]["exports"] = function(){}',
        'module[`${variable}`] = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
        'module[`${"exports"}`] = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
        'module.exports.myThing = "valid";',
        `const module = {};
module.exports = function () {};
test('a test', () => {
  expect(1).toBe(1);
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-export / valid ${index}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    })),
    // Upstream reports this, but module.export (without the final s) is an
    // ordinary property, not CommonJS module.exports. This is not an export.
    {
      name: 'upstream no-export / invalid 9',
      code: 'module.export.invalid = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    // This invalid case comes from upstream 29.16.6; 29.16.0 does not report it.
    // Assigning module["export"] does not change module.exports, so Rslint accepts
    // it. The corresponding module["exports"] assignment is checked separately.
    {
      name: 'upstream no-export / invalid 12',
      code: 'module["export"] = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    // The same local parameter is correctly ignored with type information.
    {
      name: 'typed shadowing control',
      code: "function f(module) { module.exports = {}; } test('t', () => {});",
    },
  ],
  invalid: [
    ...(
      [
        [
          'export const myThing = "invalid"; test("a test", () => { expect(1).toBe(1);});',
          1,
          1,
          34,
        ],
        [
          `
        export const myThing = 'invalid';

        test.each()('my code', () => {
          expect(1).toBe(1);
        });
     \u0020`,
          2,
          9,
          42,
        ],
        [
          `
        export const myThing = 'invalid';

        test.each\`\`('my code', () => {
          expect(1).toBe(1);
        });
     \u0020`,
          2,
          9,
          42,
        ],
        [
          `
        export const myThing = 'invalid';

        test.only.each\`\`('my code', () => {
          expect(1).toBe(1);
        });
     \u0020`,
          2,
          9,
          42,
        ],
        [
          'export default function() {};  test("a test", () => { expect(1).toBe(1);});',
          1,
          1,
          29,
        ],
        [
          'export = function() {}; test("a test", () => { expect(1).toBe(1);});',
          1,
          1,
          24,
        ],
        [
          'module.exports["invalid"] = function() {};  test("a test", () => { expect(1).toBe(1);});',
          1,
          1,
          26,
        ],
        [
          'module.exports = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
          1,
          1,
          15,
        ],
        [
          'module["exports"] = function() {}; test("a test", () => {});',
          1,
          1,
          18,
        ],
        [
          'module[`exports`].foo = function() {}; test("a test", () => {});',
          1,
          1,
          22,
        ],
        [
          'module.exports.foo.bar = function() {}; test("a test", () => {});',
          1,
          1,
          23,
        ],
        ['module.exports ||= {}; test("a test", () => {});', 1, 1, 15],
        ['value = module.exports; test("a test", () => {});', 1, 9, 23],
        ['exports.foo = "invalid"; test("a test", () => {});', 1, 1, 12],
        ['exports["foo"].bar = "invalid"; test("a test", () => {});', 1, 1, 19],
        [
          'export import foo = require("./foo"); test("a test", () => {});',
          1,
          1,
          38,
        ],
        [
          'export const myThing = "invalid"; describe("a suite", () => {});',
          1,
          1,
          34,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'unexpectedExport',
          message: 'Do not export from a test file',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          'export const myThing = "invalid"; test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 0',
          34,
        ],
        [
          `export const myThing = 'invalid';

test.each()('my code', () => {
  expect(1).toBe(1);
});`,
          'upstream no-export / invalid 1',
          34,
        ],
        [
          `export const myThing = 'invalid';

test.each\`\`('my code', () => {
  expect(1).toBe(1);
});`,
          'upstream no-export / invalid 2',
          34,
        ],
        [
          `export const myThing = 'invalid';

test.only.each\`\`('my code', () => {
  expect(1).toBe(1);
});`,
          'upstream no-export / invalid 3',
          34,
        ],
        [
          'export default function() {};  test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 4',
          29,
        ],
        [
          'export = function() {}; test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 5',
          24,
        ],
        [
          'module.exports["invalid"] = function() {};  test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 6',
          26,
        ],
        [
          'module.exports = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 7',
          15,
        ],
        [
          'module.exports.exports.exports = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 8',
          31,
        ],
        [
          'module["exports"] = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 10',
          18,
        ],
        [
          'module["exports"].invalid = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 11',
          26,
        ],
        [
          'module[`exports`] = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 13',
          18,
        ],
        [
          `module.exports = function () {};

describe('a test', () => {
  expect(1).toBe(1);
});`,
          'upstream no-export / invalid 14',
          15,
        ],
        [
          'exports["invalid"] = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 15',
          19,
        ],
        [
          'exports.export["invalid"] = function() {}; ;  test("a test", () => { expect(1).toBe(1);});',
          'upstream no-export / invalid 16',
          26,
        ],
        [
          `exports.invalid = function () {};

describe('a test', () => {
  expect(1).toBe(1);
});`,
          'upstream no-export / invalid 17',
          16,
        ],
      ] as const
    ).map(([code, name, endColumn]): TestCase => ({
      name,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedExport',
          message: 'Do not export from a test file',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    // Without a checker, the rule mistakes a local module parameter for CommonJS module.
    {
      name: 'known existing source-only shadowing defect',
      code: "function f(module) { module.exports = {}; } test('t', () => {});",
      languageOptions: {
        parserOptions: {
          project: false,
        },
      },
      errors: [
        {
          messageId: 'unexpectedExport',
          message: 'Do not export from a test file',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
  ],
});
