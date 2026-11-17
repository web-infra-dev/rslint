// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-to-have-been-called', {
  valid: [
    ...(
      [
        'expect(method.mock.calls).toHaveLength;',
        'expect(method.mock.calls).toHaveLength(0);',
        'expect(method).toHaveBeenCalledTimes(1)',
        'expect(method).not.toHaveBeenCalledTimes(x)',
        'expect(method).not.toHaveBeenCalledTimes(1)',
        'expect(method).not.toHaveBeenCalledTimes(...x)',
        'expect(a);',
        'expect(method).not.resolves.toHaveBeenCalledTimes(0);',
        'expect(method).toBeCalledTimes(0!);',
        'expect(method).toBeCalledTimes(0 satisfies number);',
        'expect(method).toBe([])',
        'expect(fn.mock.calls).toEqual([])',
        'expect(fn.mock.calls).toContain(1, 2, 3)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'expect(method.mock.calls).toHaveLength;',
        'expect(method.mock.calls).toHaveLength(0);',
        'expect(method).toHaveBeenCalledTimes(1)',
        'expect(method).not.toHaveBeenCalledTimes(x)',
        'expect(method).not.toHaveBeenCalledTimes(1)',
        'expect(method).not.toHaveBeenCalledTimes(...x)',
        'expect(a);',
        'expect(method).not.resolves.toHaveBeenCalledTimes(0);',
        'expect(method).toBe([])',
        'expect(fn.mock.calls).toEqual([])',
        'expect(fn.mock.calls).toContain(1, 2, 3)',
        'expect(method).toBeCalledTimes();',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-to-have-been-called / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          'expect(method).toBeCalledTimes(0);',
          16,
          31,
          'expect(method).not.toHaveBeenCalled();',
        ],
        [
          'expect(method).not.toBeCalledTimes(0);',
          20,
          35,
          'expect(method).toHaveBeenCalled();',
        ],
        [
          'expect(method).toHaveBeenCalledTimes(0);',
          16,
          37,
          'expect(method).not.toHaveBeenCalled();',
        ],
        [
          'expect(method).not.toHaveBeenCalledTimes(0);',
          20,
          41,
          'expect(method).toHaveBeenCalled();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 4',
      code: 'expect(method).not.toHaveBeenCalledTimes(0, 1, 2);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    ...(
      [
        [
          'expect(method).resolves.toHaveBeenCalledTimes(0);',
          25,
          46,
          'expect(method).resolves.not.toHaveBeenCalled();',
        ],
        [
          'expect(method).rejects.not.toHaveBeenCalledTimes(0);',
          28,
          49,
          'expect(method).rejects.toHaveBeenCalled();',
        ],
        [
          'expect(method).toBeCalledTimes(0 as number);',
          16,
          31,
          'expect(method).not.toHaveBeenCalled();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 5}`,
      code,
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'expect(method).toBeCalledTimes(0);',
          16,
          31,
          'expect(method).not.toHaveBeenCalled();',
        ],
        [
          'expect(method).not.toBeCalledTimes(0);',
          20,
          35,
          'expect(method).toHaveBeenCalled();',
        ],
        [
          'expect(method).toHaveBeenCalledTimes(0);',
          16,
          37,
          'expect(method).not.toHaveBeenCalled();',
        ],
        [
          'expect(method).not.toHaveBeenCalledTimes(0);',
          20,
          41,
          'expect(method).toHaveBeenCalled();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-to-have-been-called / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-to-have-been-called / invalid 4',
      code: 'expect(method).not.toHaveBeenCalledTimes(0, 1, 2);',
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    ...(
      [
        [
          'expect(method).resolves.toHaveBeenCalledTimes(0);',
          25,
          46,
          'expect(method).resolves.not.toHaveBeenCalled();',
        ],
        [
          'expect(method).rejects.not.toHaveBeenCalledTimes(0);',
          28,
          49,
          'expect(method).rejects.toHaveBeenCalled();',
        ],
        [
          'expect(method).toBeCalledTimes(0 as number);',
          16,
          31,
          'expect(method).not.toHaveBeenCalled();',
        ],
      ] as const
    ).map(([code, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-to-have-been-called / invalid ${index + 5}`,
      code,
      languageOptions: {
        ecmaVersion: 2020,
      },
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
  ],
});
