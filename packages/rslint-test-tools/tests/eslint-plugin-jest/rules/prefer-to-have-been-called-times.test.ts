import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-to-have-been-called-times', {} as never, {
  valid: [
    { code: 'expect.assertions(1)' },
    { code: 'expect(fn).toHaveBeenCalledTimes' },
    { code: 'expect(fn.mock.calls).toHaveLength' },
    { code: 'expect(fn.mock.values).toHaveLength(0)' },
    { code: 'expect(fn.values.calls).toHaveLength(0)' },
    { code: 'expect(fn).toHaveBeenCalledTimes(0)' },
    { code: 'expect(fn).resolves.toHaveBeenCalledTimes(10)' },
    { code: 'expect(fn).not.toHaveBeenCalledTimes(10)' },
    { code: 'expect(fn).toHaveBeenCalledTimes(1)' },
    { code: 'expect(fn).toBeCalledTimes(0);' },
    { code: 'expect(fn).toHaveBeenCalledTimes(0);' },
    { code: 'expect(fn);' },
    { code: 'expect(method.mock.calls[0][0]).toStrictEqual(value);' },
    { code: 'expect(fn.mock.length).toEqual(1);' },
    { code: 'expect(fn.mock.calls).toEqual([]);' },
    { code: 'expect(fn.mock.calls).toContain(1, 2, 3);' },
    { code: 'expect((fn.mock.calls)).toEqual([]);' },
  ],
  invalid: [
    {
      code: 'expect(method.mock.calls).toHaveLength(1);',
      output: 'expect(method).toHaveBeenCalledTimes(1);',
      errors: [
        {
          messageId: 'preferMatcher',
          column: 27,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(method.mock.calls).resolves.toHaveLength(x);',
      output: 'expect(method).resolves.toHaveBeenCalledTimes(x);',
      errors: [
        {
          messageId: 'preferMatcher',
          column: 36,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(method["mock"].calls).toHaveLength(0);',
      output: 'expect(method).toHaveBeenCalledTimes(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          column: 30,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(my.method.mock.calls).not.toHaveLength(0);',
      output: 'expect(my.method).not.toHaveBeenCalledTimes(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          column: 34,
          line: 1,
        },
      ],
    },
    {
      code: 'expect((method.mock.calls)).toHaveLength(1);',
      output: 'expect(method).toHaveBeenCalledTimes(1);',
      errors: [
        {
          messageId: 'preferMatcher',
          column: 29,
          line: 1,
        },
      ],
    },
  ],
});
