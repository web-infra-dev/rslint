// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-restricted-matchers', {
  valid: [
    ...(
      [
        'expect(a).toHaveBeenCalled()',
        'expect(a).not.toHaveBeenCalled()',
        'expect(a).toHaveBeenCalledTimes()',
        'expect(a).toHaveBeenCalledWith()',
        'expect(a).toHaveBeenLastCalledWith()',
        'expect(a).toHaveBeenNthCalledWith()',
        'expect(a).toHaveReturned()',
        'expect(a).toHaveReturnedTimes()',
        'expect(a).toHaveReturnedWith()',
        'expect(a).toHaveLastReturnedWith()',
        'expect(a).toHaveNthReturnedWith()',
        'expect(a).toThrow()',
        'expect(a).rejects;',
        'expect(a);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 14',
      code: 'expect(a).resolves',
      options: [
        {
          not: null,
        },
      ],
    },
    {
      name: 'run 0 valid 15',
      code: 'expect(a).toBe(b)',
      options: [
        {
          'not.toBe': null,
        },
      ],
    },
    {
      name: 'run 0 valid 16',
      code: 'expect(a).toBeUndefined(b)',
      options: [
        {
          toBe: null,
        },
      ],
    },
    {
      name: 'run 0 valid 17',
      code: 'expect(a)["toBe"](b)',
      options: [
        {
          'not.toBe': null,
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: 'expect(a).resolves.not.toBe(b)',
      options: [
        {
          not: null,
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: 'expect(a).resolves.not.toBe(b)',
      options: [
        {
          'not.toBe': null,
        },
      ],
    },
    ...(
      [
        "expect(uploadFileMock).resolves.toHaveBeenCalledWith('file.name')",
        "expect(uploadFileMock).resolves.not.toHaveBeenCalledWith('file.name')",
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 20}`,
      code,
      options: [
        {
          'not.toHaveBeenCalledWith': 'Use not.toHaveBeenCalled instead',
        },
      ],
    })),
    ...(
      [
        'expect(a).toHaveBeenCalled()',
        'expect(a).not.toHaveBeenCalled()',
        'expect(a).toHaveBeenCalledTimes()',
        'expect(a).toHaveBeenCalledWith()',
        'expect(a).toHaveBeenLastCalledWith()',
        'expect(a).toHaveBeenNthCalledWith()',
        'expect(a).toHaveReturned()',
        'expect(a).toHaveReturnedTimes()',
        'expect(a).toHaveReturnedWith()',
        'expect(a).toHaveLastReturnedWith()',
        'expect(a).toHaveNthReturnedWith()',
        'expect(a).toThrow()',
        'expect(a).rejects;',
        'expect(a);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-restricted-matchers / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
    {
      name: 'upstream no-restricted-matchers / valid 14',
      code: 'expect(a).resolves',
      options: [
        {
          not: null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream no-restricted-matchers / valid 15',
      code: 'expect(a).toBe(b)',
      options: [
        {
          'not.toBe': null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream no-restricted-matchers / valid 16',
      code: 'expect(a).toBeUndefined(b)',
      options: [
        {
          toBe: null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream no-restricted-matchers / valid 17',
      code: 'expect(a)["toBe"](b)',
      options: [
        {
          'not.toBe': null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream no-restricted-matchers / valid 18',
      code: 'expect(a).resolves.not.toBe(b)',
      options: [
        {
          not: null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      name: 'upstream no-restricted-matchers / valid 19',
      code: 'expect(a).resolves.not.toBe(b)',
      options: [
        {
          'not.toBe': null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    ...(
      [
        "expect(uploadFileMock).resolves.toHaveBeenCalledWith('file.name')",
        "expect(uploadFileMock).resolves.not.toHaveBeenCalledWith('file.name')",
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-restricted-matchers / valid ${index + 20}`,
      code,
      options: [
        {
          'not.toHaveBeenCalledWith': 'Use not.toHaveBeenCalled instead',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    ...(
      [
        ['expect(a).toBe(b)', 15],
        ['expect(a)["toBe"](b)', 17],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      options: [
        {
          toBe: null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `toBe` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(a).not[x]()', 16],
        ['expect(a).not.toBe(b)', 19],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 2}`,
      code,
      options: [
        {
          not: null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `not` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(a).resolves.toBe(b)', 24],
        ['expect(a).resolves.not.toBe(b)', 28],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 4}`,
      code,
      options: [
        {
          resolves: null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `resolves` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 6',
      code: 'expect(a).resolves.not.toBe(b)',
      options: [
        {
          'resolves.not': null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `resolves.not` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'expect(a).not.toBe(b)',
      options: [
        {
          'not.toBe': null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `not.toBe` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'expect(a).resolves.not.toBe(b)',
      options: [
        {
          'resolves.not.toBe': null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `resolves.not.toBe` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'expect(a).toBe(b)',
      options: [
        {
          toBe: 'Prefer `toStrictEqual` instead',
        },
      ],
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          message: 'Prefer `toStrictEqual` instead',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: `
        test('some test', async () => {
          await expect(Promise.resolve(1)).resolves.toBe(1);
         });
     \u0020`,
      options: [
        {
          resolves: 'Use `expect(await promise)` instead.',
        },
      ],
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          message: 'Use `expect(await promise)` instead.',
          line: 3,
          column: 44,
          endLine: 3,
          endColumn: 57,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'expect(Promise.resolve({})).rejects.toBeFalsy()',
      options: [
        {
          'rejects.toBeFalsy': null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `rejects.toBeFalsy` is disallowed',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: "expect(uploadFileMock).not.toHaveBeenCalledWith('file.name')",
      options: [
        {
          'not.toHaveBeenCalledWith': 'Use not.toHaveBeenCalled instead',
        },
      ],
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          message: 'Use not.toHaveBeenCalled instead',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'expect(a). /* c */ toBe(b)',
      options: [
        {
          toBe: null,
        },
      ],
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `toBe` is disallowed',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    ...(
      [
        ['expect(a).toBe(b)', 15],
        ['expect(a)["toBe"](b)', 17],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-restricted-matchers / invalid ${index}`,
      code,
      options: [
        {
          toBe: null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `toBe` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(a).not[x]()', 16],
        ['expect(a).not.toBe(b)', 19],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-restricted-matchers / invalid ${index + 2}`,
      code,
      options: [
        {
          not: null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `not` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['expect(a).resolves.toBe(b)', 24],
        ['expect(a).resolves.not.toBe(b)', 28],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-restricted-matchers / invalid ${index + 4}`,
      code,
      options: [
        {
          resolves: null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `resolves` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-restricted-matchers / invalid 6',
      code: 'expect(a).resolves.not.toBe(b)',
      options: [
        {
          'resolves.not': null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `resolves.not` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'upstream no-restricted-matchers / invalid 7',
      code: 'expect(a).not.toBe(b)',
      options: [
        {
          'not.toBe': null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `not.toBe` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'upstream no-restricted-matchers / invalid 8',
      code: 'expect(a).resolves.not.toBe(b)',
      options: [
        {
          'resolves.not.toBe': null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `resolves.not.toBe` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'upstream no-restricted-matchers / invalid 9',
      code: 'expect(a).toBe(b)',
      options: [
        {
          toBe: 'Prefer `toStrictEqual` instead',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          message: 'Prefer `toStrictEqual` instead',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'upstream no-restricted-matchers / invalid 10',
      code: `
        test('some test', async () => {
          await expect(Promise.resolve(1)).resolves.toBe(1);
         });
     \u0020`,
      options: [
        {
          resolves: 'Use `expect(await promise)` instead.',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          message: 'Use `expect(await promise)` instead.',
          line: 3,
          column: 44,
          endLine: 3,
          endColumn: 57,
        },
      ],
    },
    {
      name: 'upstream no-restricted-matchers / invalid 11',
      code: 'expect(Promise.resolve({})).rejects.toBeFalsy()',
      options: [
        {
          'rejects.toBeFalsy': null,
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `rejects.toBeFalsy` is disallowed',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'upstream no-restricted-matchers / invalid 12',
      code: "expect(uploadFileMock).not.toHaveBeenCalledWith('file.name')",
      options: [
        {
          'not.toHaveBeenCalledWith': 'Use not.toHaveBeenCalled instead',
        },
      ],
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          message: 'Use not.toHaveBeenCalled instead',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
  ],
});
