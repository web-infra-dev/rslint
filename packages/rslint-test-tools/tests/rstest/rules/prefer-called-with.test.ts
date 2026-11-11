import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-called-with', {
  valid: [
    { code: 'expect(fn).toBeCalledWith();' },
    { code: 'expect(fn).toHaveBeenCalledWith();' },
    { code: 'expect(fn).toBeCalledWith(expect.anything());' },
    { code: 'expect(fn).toHaveBeenCalledWith(expect.anything());' },
    { code: 'expect(fn).not.toBeCalled();' },
    { code: 'expect(fn).rejects.not.toBeCalled();' },
    { code: 'expect(fn).not.toHaveBeenCalled();' },
    { code: 'expect(fn).not.toBeCalledWith();' },
    { code: 'expect(fn).not.toHaveBeenCalledWith();' },
    { code: 'expect(fn).resolves.not.toHaveBeenCalledWith();' },
    { code: 'expect(fn).toBeCalledTimes(0);' },
    { code: 'expect(fn).toHaveBeenCalledTimes(0);' },
    { code: 'expect(fn);' },
    { code: 'expect(fn).toHaveBeenCalledExactlyOnceWith()' },
  ],
  invalid: [
    {
      code: 'expect(fn).toBeCalled();',
      output: 'expect(fn).toBeCalledWith();',
      errors: [
        {
          messageId: 'preferCalledWith',
          message: 'Prefer toBeCalledWith(/* expected args */)',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'expect(fn).resolves.toBeCalled();',
      output: 'expect(fn).resolves.toBeCalledWith();',
      errors: [
        {
          messageId: 'preferCalledWith',
          message: 'Prefer toBeCalledWith(/* expected args */)',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'expect(fn).toHaveBeenCalled();',
      output: 'expect(fn).toHaveBeenCalledWith();',
      errors: [
        {
          messageId: 'preferCalledWith',
          message: 'Prefer toHaveBeenCalledWith(/* expected args */)',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'it("some test", () => {expect(mockApi).toHaveBeenCalledOnce();});',
      output:
        'it("some test", () => {expect(mockApi).toHaveBeenCalledExactlyOnceWith();});',
      errors: [
        {
          messageId: 'preferCalledWith',
          message:
            'Prefer toHaveBeenCalledExactlyOnceWith(/* expected args */)',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
  ],
});
