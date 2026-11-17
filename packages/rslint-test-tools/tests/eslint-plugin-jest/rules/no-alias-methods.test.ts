// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-alias-methods', {
  valid: [
    ...(
      [
        'expect(a).toHaveBeenCalled()',
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
        `const toBeCalled = 'toHaveBeenCalled';
expect(a)[toBeCalled]();`,
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
          'expect(a).toBeCalled()',
          'Replace toBeCalled() with its canonical name of toHaveBeenCalled()',
          1,
          11,
          21,
          'expect(a).toHaveBeenCalled()',
        ],
        [
          'expect(a).toBeCalledTimes()',
          'Replace toBeCalledTimes() with its canonical name of toHaveBeenCalledTimes()',
          1,
          11,
          26,
          'expect(a).toHaveBeenCalledTimes()',
        ],
        [
          'expect(a).toBeCalledWith()',
          'Replace toBeCalledWith() with its canonical name of toHaveBeenCalledWith()',
          1,
          11,
          25,
          'expect(a).toHaveBeenCalledWith()',
        ],
        [
          'expect(a).lastCalledWith()',
          'Replace lastCalledWith() with its canonical name of toHaveBeenLastCalledWith()',
          1,
          11,
          25,
          'expect(a).toHaveBeenLastCalledWith()',
        ],
        [
          'expect(a).nthCalledWith()',
          'Replace nthCalledWith() with its canonical name of toHaveBeenNthCalledWith()',
          1,
          11,
          24,
          'expect(a).toHaveBeenNthCalledWith()',
        ],
        [
          'expect(a).toReturn()',
          'Replace toReturn() with its canonical name of toHaveReturned()',
          1,
          11,
          19,
          'expect(a).toHaveReturned()',
        ],
        [
          'expect(a).toReturnTimes()',
          'Replace toReturnTimes() with its canonical name of toHaveReturnedTimes()',
          1,
          11,
          24,
          'expect(a).toHaveReturnedTimes()',
        ],
        [
          'expect(a).toReturnWith()',
          'Replace toReturnWith() with its canonical name of toHaveReturnedWith()',
          1,
          11,
          23,
          'expect(a).toHaveReturnedWith()',
        ],
        [
          'expect(a).lastReturnedWith()',
          'Replace lastReturnedWith() with its canonical name of toHaveLastReturnedWith()',
          1,
          11,
          27,
          'expect(a).toHaveLastReturnedWith()',
        ],
        [
          'expect(a).nthReturnedWith()',
          'Replace nthReturnedWith() with its canonical name of toHaveNthReturnedWith()',
          1,
          11,
          26,
          'expect(a).toHaveNthReturnedWith()',
        ],
        [
          'expect(a).toThrowError()',
          'Replace toThrowError() with its canonical name of toThrow()',
          1,
          11,
          23,
          'expect(a).toThrow()',
        ],
        [
          'expect(a).resolves.toThrowError()',
          'Replace toThrowError() with its canonical name of toThrow()',
          1,
          20,
          32,
          'expect(a).resolves.toThrow()',
        ],
        [
          'expect(a).rejects.toThrowError()',
          'Replace toThrowError() with its canonical name of toThrow()',
          1,
          19,
          31,
          'expect(a).rejects.toThrow()',
        ],
        [
          'expect(a).not.toThrowError()',
          'Replace toThrowError() with its canonical name of toThrow()',
          1,
          15,
          27,
          'expect(a).not.toThrow()',
        ],
        [
          'expect(a).not["toThrowError"]()',
          'Replace toThrowError() with its canonical name of toThrow()',
          1,
          15,
          29,
          'expect(a).not["toThrow"]()',
        ],
        [
          `expect(a).
  toBeCalled()`,
          'Replace toBeCalled() with its canonical name of toHaveBeenCalled()',
          2,
          3,
          13,
          `expect(a).
  toHaveBeenCalled()`,
        ],
        [
          'expect(a) . /* c */ toBeCalled()',
          'Replace toBeCalled() with its canonical name of toHaveBeenCalled()',
          1,
          21,
          31,
          'expect(a) . /* c */ toHaveBeenCalled()',
        ],
        [
          "expect(a)[/* c */ 'toBeCalled']()",
          'Replace toBeCalled() with its canonical name of toHaveBeenCalled()',
          1,
          19,
          31,
          "expect(a)[/* c */ 'toHaveBeenCalled']()",
        ],
      ] as const
    ).map(
      ([code, message, line, column, endColumn, output], index): TestCase => ({
        name: `run 0 invalid ${index}`,
        code,
        errors: [
          {
            messageId: 'replaceAlias',
            message,
            line,
            column,
            endLine: line,
            endColumn,
          },
        ],
        output,
      }),
    ),
  ],
});
