// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-standalone-expect', {
  valid: [
    ...(
      [
        'expect.any(String)',
        'expect.extend({})',
        'describe("a test", () => { it("an it", () => {expect(1).toBe(1); }); });',
        'describe("a test", () => { it("an it", () => { const func = () => { expect(1).toBe(1); }; }); });',
        'describe("a test", () => { const func = () => { expect(1).toBe(1); }; });',
        'describe("a test", () => { function func() { expect(1).toBe(1); }; });',
        'describe("a test", () => { const func = function(){ expect(1).toBe(1); }; });',
        'it("an it", () => expect(1).toBe(1))',
        'const func = function(){ expect(1).toBe(1); };',
        'const func = () => expect(1).toBe(1);',
        '{}',
        'it.each([1, true])("trues", value => { expect(value).toBe(true); });',
        'it.each([1, true])("trues", value => { expect(value).toBe(true); }); it("an it", () => { expect(1).toBe(1) });',
        `
      it.each\`
        num   | value
        \${1} | \${true}
      \`('trues', ({ value }) => {
        expect(value).toBe(true);
      });
   \u0020`,
        'it.only("an only", value => { expect(value).toBe(true); });',
        'it.concurrent("an concurrent", value => { expect(value).toBe(true); });',
        'describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });',
        'describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });',
        'const helpers = { assert() { expect(1).toBe(1); } };',
        'class Helper { assert() { expect(1).toBe(1); } get value() { expect(1).toBe(1); return 1; } }',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 20',
      code: `
        describe('scenario', () => {
          const t = Math.random() ? it.only : it;
          t('testing', () => expect(true));
        });
     \u0020`,
      options: [
        {
          additionalTestBlockFunctions: ['t'],
        },
      ],
    },
    {
      name: 'run 0 valid 21',
      code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
     \u0020`,
      options: [
        {
          additionalTestBlockFunctions: ['each.test'],
        },
      ],
    },
    ...(
      [
        'expect.any(String)',
        'expect.extend({})',
        'describe("a test", () => { it("an it", () => {expect(1).toBe(1); }); });',
        'describe("a test", () => { it("an it", () => { const func = () => { expect(1).toBe(1); }; }); });',
        'describe("a test", () => { const func = () => { expect(1).toBe(1); }; });',
        'describe("a test", () => { function func() { expect(1).toBe(1); }; });',
        'describe("a test", () => { const func = function(){ expect(1).toBe(1); }; });',
        'it("an it", () => expect(1).toBe(1))',
        'const func = function(){ expect(1).toBe(1); };',
        'const func = () => expect(1).toBe(1);',
        '{}',
        'it.each([1, true])("trues", value => { expect(value).toBe(true); });',
        'it.each([1, true])("trues", value => { expect(value).toBe(true); }); it("an it", () => { expect(1).toBe(1) });',
        `it.each\`
  num   | value
  \${1} | \${true}
\`('trues', ({ value }) => {
  expect(value).toBe(true);
});`,
        'it.only("an only", value => { expect(value).toBe(true); });',
        'it.concurrent("an concurrent", value => { expect(value).toBe(true); });',
        'describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-standalone-expect / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream no-standalone-expect / valid 17',
      code: `describe('scenario', () => {
  const t = Math.random() ? it.only : it;
  t('testing', () => expect(true));
});`,
      options: [
        {
          additionalTestBlockFunctions: ['t'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream no-standalone-expect / valid 18',
      code: `each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
]).test('returns the result of adding %d to %d', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
      options: [
        {
          additionalTestBlockFunctions: ['each.test'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
  ],
  invalid: [
    ...(
      [
        ["(() => {})('testing', () => expect(true).toBe(false))", 1, 29, 53],
        ['expect.hasAssertions()', 1, 1, 23],
        ['expect().hasAssertions()', 1, 1, 25],
        [
          `
        describe('scenario', () => {
          const t = Math.random() ? it.only : it;
          t('testing', () => expect(true).toBe(false));
        });
     \u0020`,
          4,
          30,
          54,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 4',
      code: `
        describe('scenario', () => {
          const t = Math.random() ? it.only : it;
          t('testing', () => expect(true).toBe(false));
        });
     \u0020`,
      options: [
        {
          additionalTestBlockFunctions: undefined,
        },
      ],
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 4,
          column: 30,
          endLine: 4,
          endColumn: 54,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
     \u0020`,
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
     \u0020`,
      options: [
        {
          additionalTestBlockFunctions: ['each'],
        },
      ],
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
     \u0020`,
      options: [
        {
          additionalTestBlockFunctions: ['test'],
        },
      ],
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 39,
        },
      ],
    },
    ...(
      [
        ['describe("a test", () => { expect(1).toBe(1); });', 1, 28, 45],
        ['describe("a test", () => expect(1).toBe(1));', 1, 26, 43],
        [
          'describe("a test", () => { const func = () => { expect(1).toBe(1); }; expect(1).toBe(1); });',
          1,
          71,
          88,
        ],
        [
          'describe("a test", () => {  it(() => { expect(1).toBe(1); }); expect(1).toBe(1); });',
          1,
          63,
          80,
        ],
        ['expect(1).toBe(1);', 1, 1, 18],
        ['{expect(1).toBe(1)}', 1, 2, 19],
        [
          'it.each([1, true])("trues", value => { expect(value).toBe(true); }); expect(1).toBe(1);',
          1,
          70,
          87,
        ],
        [
          'describe.each([1, true])("trues", value => { expect(value).toBe(true); });',
          1,
          46,
          70,
        ],
        [
          `
        describe.each\`
          num   | value
          \${1} | \${true}
        \`('trues', ({ value }) => {
          expect(value).toBe(true);
        });
     \u0020`,
          6,
          11,
          35,
        ],
        [
          `
        it.each\`
          num   | value
          \${1} | \${true}
        \`('trues', ({ value }) => {
          expect(value).toBe(true);
        });

        expect(1).toBe(1);
     \u0020`,
          9,
          9,
          26,
        ],
        [
          `
        import { expect as pleaseExpect } from '@jest/globals';

        describe("a test", () => { pleaseExpect(1).toBe(1); });
     \u0020`,
          4,
          36,
          59,
        ],
        [
          `
        import { expect as pleaseExpect } from '@jest/globals';

        beforeEach(() => pleaseExpect.hasAssertions());
     \u0020`,
          4,
          26,
          54,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 8}`,
      code,
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ["(() => {})('testing', () => expect(true).toBe(false))", 1, 29, 53],
        ['expect.hasAssertions()', 1, 1, 23],
        ['expect().hasAssertions()', 1, 1, 25],
        [
          `describe('scenario', () => {
  const t = Math.random() ? it.only : it;
  t('testing', () => expect(true).toBe(false));
});`,
          3,
          22,
          46,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `upstream no-standalone-expect / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-standalone-expect / invalid 4',
      code: `describe('scenario', () => {
  const t = Math.random() ? it.only : it;
  t('testing', () => expect(true).toBe(false));
});`,
      options: [
        {
          additionalTestBlockFunctions: undefined,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 3,
          column: 22,
          endLine: 3,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream no-standalone-expect / invalid 5',
      code: `each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
]).test('returns the result of adding %d to %d', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'upstream no-standalone-expect / invalid 6',
      code: `each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
]).test('returns the result of adding %d to %d', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
      options: [
        {
          additionalTestBlockFunctions: ['each'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'upstream no-standalone-expect / invalid 7',
      code: `each([
  [1, 1, 2],
  [1, 2, 3],
  [2, 1, 3],
]).test('returns the result of adding %d to %d', (a, b, expected) => {
  expect(a + b).toBe(expected);
});`,
      options: [
        {
          additionalTestBlockFunctions: ['test'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 6,
          column: 3,
          endLine: 6,
          endColumn: 31,
        },
      ],
    },
    ...(
      [
        ['describe("a test", () => { expect(1).toBe(1); });', 28, 45],
        ['describe("a test", () => expect(1).toBe(1));', 26, 43],
        [
          'describe("a test", () => { const func = () => { expect(1).toBe(1); }; expect(1).toBe(1); });',
          71,
          88,
        ],
        [
          'describe("a test", () => {  it(() => { expect(1).toBe(1); }); expect(1).toBe(1); });',
          63,
          80,
        ],
        ['expect(1).toBe(1);', 1, 18],
        ['{expect(1).toBe(1)}', 2, 19],
        [
          'it.each([1, true])("trues", value => { expect(value).toBe(true); }); expect(1).toBe(1);',
          70,
          87,
        ],
        [
          'describe.each([1, true])("trues", value => { expect(value).toBe(true); });',
          46,
          70,
        ],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream no-standalone-expect / invalid ${index + 8}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `import { expect as pleaseExpect } from '@jest/globals';

describe("a test", () => { pleaseExpect(1).toBe(1); });`,
          28,
          51,
        ],
        [
          `import { expect as pleaseExpect } from '@jest/globals';

beforeEach(() => pleaseExpect.hasAssertions());`,
          18,
          46,
        ],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `upstream no-standalone-expect / invalid ${index + 16}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedExpect',
          message: 'Expect must be inside of a test block',
          line: 3,
          column,
          endLine: 3,
          endColumn,
        },
      ],
    })),
  ],
});
