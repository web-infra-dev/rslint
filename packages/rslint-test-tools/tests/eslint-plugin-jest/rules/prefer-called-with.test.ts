// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-called-with', {
  valid: [
    ...(
      [
        'expect(fn).toBeCalledWith();',
        'expect(fn).toHaveBeenCalledWith();',
        'expect(fn).toBeCalledWith(expect.anything());',
        'expect(fn).toHaveBeenCalledWith(expect.anything());',
        'expect(fn).not.toBeCalled();',
        'expect(fn).rejects.not.toBeCalled();',
        'expect(fn).not.toHaveBeenCalled();',
        'expect(fn).not.toBeCalledWith();',
        'expect(fn).not.toHaveBeenCalledWith();',
        'expect(fn).resolves.not.toHaveBeenCalledWith();',
        'expect(fn).toBeCalledTimes(0);',
        'expect(fn).toHaveBeenCalledTimes(0);',
        'expect(fn);',
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
          'expect(fn).toBeCalled();',
          'Prefer toBeCalledWith(/* expected args */)',
          12,
          22,
        ],
        [
          'expect(fn).resolves.toBeCalled();',
          'Prefer toBeCalledWith(/* expected args */)',
          21,
          31,
        ],
        [
          'expect(fn).toHaveBeenCalled();',
          'Prefer toHaveBeenCalledWith(/* expected args */)',
          12,
          28,
        ],
      ] as const
    ).map(([code, message, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'preferCalledWith',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
  ],
});
