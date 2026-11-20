// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-to-have-been-called-times', {
  valid: [
    ...(
      [
        'expect.assertions(1)',
        'expect(fn).toHaveBeenCalledTimes',
        'expect(fn.mock.calls).toHaveLength',
        'expect(fn.mock.values).toHaveLength(0)',
        'expect(fn.values.calls).toHaveLength(0)',
        'expect(fn).toHaveBeenCalledTimes(0)',
        'expect(fn).resolves.toHaveBeenCalledTimes(10)',
        'expect(fn).not.toHaveBeenCalledTimes(10)',
        'expect(fn).toHaveBeenCalledTimes(1)',
        'expect(fn).toBeCalledTimes(0);',
        'expect(fn).toHaveBeenCalledTimes(0);',
        'expect(fn);',
        'expect(method.mock.calls[0][0]).toStrictEqual(value);',
        'expect(fn.mock.length).toEqual(1);',
        'expect(fn.mock.calls).toEqual([]);',
        'expect(fn.mock.calls).toContain(1, 2, 3);',
        'expect((fn.mock.calls)).toEqual([]);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
  ],
  invalid: [
    ...(
      [
        [
          'expect(method.mock.calls).toHaveLength(1);',
          27,
          'expect(method).toHaveBeenCalledTimes(1);',
        ],
        [
          'expect(method.mock.calls).resolves.toHaveLength(x);',
          36,
          'expect(method).resolves.toHaveBeenCalledTimes(x);',
        ],
        [
          'expect(method["mock"].calls).toHaveLength(0);',
          30,
          'expect(method).toHaveBeenCalledTimes(0);',
        ],
        [
          'expect(my.method.mock.calls).not.toHaveLength(0);',
          34,
          'expect(my.method).not.toHaveBeenCalledTimes(0);',
        ],
        [
          'expect((method.mock.calls)).toHaveLength(1);',
          29,
          'expect((method)).toHaveBeenCalledTimes(1);',
        ],
      ] as const
    ).map(([code, column, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 12,
        },
      ],
      output,
    })),
  ],
});
