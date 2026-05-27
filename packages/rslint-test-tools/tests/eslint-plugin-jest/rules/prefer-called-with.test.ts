import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-called-with', {} as never, {
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
  ],
  invalid: [
    {
      code: 'expect(fn).toBeCalled();',
      errors: [
        {
          messageId: 'preferCalledWith',
          data: { matcherName: 'toBeCalled' },
          column: 12,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(fn).resolves.toBeCalled();',
      errors: [
        {
          messageId: 'preferCalledWith',
          data: { matcherName: 'toBeCalled' },
          column: 21,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(fn).toHaveBeenCalled();',
      errors: [
        {
          messageId: 'preferCalledWith',
          data: { matcherName: 'toHaveBeenCalled' },
          column: 12,
          line: 1,
        },
      ],
    },
  ],
});
