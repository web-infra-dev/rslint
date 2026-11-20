// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-hooks', {
  valid: [
    ...(
      [
        'test("foo")',
        'describe("foo", () => { it("bar") })',
        'test("foo", () => { expect(subject.beforeEach()).toBe(true) })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 3',
      code: 'afterEach(() => {}); afterAll(() => {});',
      options: [
        {
          allow: ['afterEach', 'afterAll'],
        },
      ],
    },
    ...(
      [
        'test("foo")',
        'describe("foo", () => { it("bar") })',
        'test("foo", () => { expect(subject.beforeEach()).toBe(true) })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-hooks / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream no-hooks / valid 3',
      code: 'afterEach(() => {}); afterAll(() => {});',
      options: [
        {
          allow: ['afterEach', 'afterAll'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream no-hooks / valid 4',
      code: 'test("foo")',
      options: [
        {
          allow: undefined,
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
        ['beforeAll(() => {})', "Unexpected 'beforeAll' hook", 1, 1, 20],
        ['beforeEach(() => {})', "Unexpected 'beforeEach' hook", 1, 1, 21],
        ['afterAll(() => {})', "Unexpected 'afterAll' hook", 1, 1, 19],
        ['afterEach(() => {})', "Unexpected 'afterEach' hook", 1, 1, 20],
        [
          `
        import { 'afterEach' as afterEachTest } from '@jest/globals';

        afterEachTest(() => {})
     \u0020`,
          "Unexpected 'afterEach' hook",
          4,
          9,
          32,
        ],
      ] as const
    ).map(([code, message, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'unexpectedHook',
          message,
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
          'beforeEach(() => {}); afterEach(() => { jest.resetModules() });',
          1,
          1,
          21,
        ],
        [
          `
        import { beforeEach as afterEach, afterEach as beforeEach } from '@jest/globals';

        afterEach(() => {});
        beforeEach(() => { jest.resetModules() });
     \u0020`,
          4,
          9,
          28,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 5}`,
      code,
      options: [
        {
          allow: ['afterEach'],
        },
      ],
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'beforeEach' hook",
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['beforeAll(() => {})', "Unexpected 'beforeAll' hook", 20],
        ['beforeEach(() => {})', "Unexpected 'beforeEach' hook", 21],
        ['afterAll(() => {})', "Unexpected 'afterAll' hook", 19],
        ['afterEach(() => {})', "Unexpected 'afterEach' hook", 20],
      ] as const
    ).map(([code, message, endColumn], index): TestCase => ({
      name: `upstream no-hooks / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedHook',
          message,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-hooks / invalid 4',
      code: `import { 'afterEach' as afterEachTest } from '@jest/globals';

afterEachTest(() => {})`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2022,
      },
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'afterEach' hook",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'upstream no-hooks / invalid 5',
      code: 'beforeEach(() => {}); afterEach(() => { jest.resetModules() });',
      options: [
        {
          allow: ['afterEach'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'beforeEach' hook",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'upstream no-hooks / invalid 6',
      code: `import { beforeEach as afterEach, afterEach as beforeEach } from '@jest/globals';

afterEach(() => {});
beforeEach(() => { jest.resetModules() });`,
      options: [
        {
          allow: ['afterEach'],
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'beforeEach' hook",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
  ],
});
