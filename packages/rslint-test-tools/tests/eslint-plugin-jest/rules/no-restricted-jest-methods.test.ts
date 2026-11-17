// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-restricted-jest-methods', {
  valid: [
    ...(
      [
        'jest',
        'jest()',
        'jest.mock()',
        'expect(a).rejects;',
        'expect(a);',
        `
        import { jest } from '@jest/globals';

        jest;
     \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'jest',
        'jest()',
        'jest.mock()',
        'expect(a).rejects;',
        'expect(a);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-restricted-jest-methods / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream no-restricted-jest-methods / valid 5',
      code: `import { jest } from '@jest/globals';

jest;`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
    },
  ],
  invalid: [
    ...(
      [
        ['jest.fn()', 8],
        ['jest["fn"]()', 10],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      options: [
        {
          fn: null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedJestMethod',
          message: 'Use of `fn` is disallowed',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['jest.mock()', 10],
        ['jest["mock"]()', 12],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 2}`,
      code,
      options: [
        {
          mock: 'Do not use mocks',
        },
      ],
      errors: [
        {
          messageId: 'restrictedJestMethodWithMessage',
          message: 'Do not use mocks',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `
        import { jest } from '@jest/globals';

        jest.advanceTimersByTime();
     \u0020`,
          4,
          14,
        ],
        ['jest. /* c */ advanceTimersByTime();', 1, 15],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `run 0 invalid ${index + 4}`,
      code,
      options: [
        {
          advanceTimersByTime: null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedJestMethod',
          message: 'Use of `advanceTimersByTime` is disallowed',
          line,
          column,
          endLine: line,
          endColumn: column + 19,
        },
      ],
    })),
    ...(
      [
        ['jest.fn()', 8],
        ['jest["fn"]()', 10],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-restricted-jest-methods / invalid ${index}`,
      code,
      options: [
        {
          fn: null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedJestMethod',
          message: 'Use of `fn` is disallowed',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['jest.mock()', 10],
        ['jest["mock"]()', 12],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-restricted-jest-methods / invalid ${index + 2}`,
      code,
      options: [
        {
          mock: 'Do not use mocks',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedJestMethodWithMessage',
          message: 'Do not use mocks',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-restricted-jest-methods / invalid 4',
      code: `import { jest } from '@jest/globals';

jest.advanceTimersByTime();`,
      options: [
        {
          advanceTimersByTime: null,
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedJestMethod',
          message: 'Use of `advanceTimersByTime` is disallowed',
          line: 3,
          column: 6,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
  ],
});
