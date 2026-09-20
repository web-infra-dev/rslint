import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-to-have-been-called-times', {} as never, {
  valid: [
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
    'expect(fn?.mock.calls).toHaveLength(1);',
    'expect(fn[mock].calls).toHaveLength(1);',
  ].map((code) => ({ code })),
  invalid: [
    {
      code: 'expect(method.mock.calls).toHaveLength(1);',
      output: 'expect(method).toHaveBeenCalledTimes(1);',
      column: 27,
      endColumn: 39,
    },
    {
      code: 'expect(method.mock.calls).resolves.toHaveLength(x);',
      output: 'expect(method).resolves.toHaveBeenCalledTimes(x);',
      column: 36,
      endColumn: 48,
    },
    {
      code: 'expect(method["mock"].calls).toHaveLength(0);',
      output: 'expect(method).toHaveBeenCalledTimes(0);',
      column: 30,
      endColumn: 42,
    },
    {
      code: 'expect(my.method.mock.calls).not.toHaveLength(0);',
      output: 'expect(my.method).not.toHaveBeenCalledTimes(0);',
      column: 34,
      endColumn: 46,
    },
    {
      code: 'expect.soft(fn.mock.calls).toHaveLength(2);',
      output: 'expect.soft(fn).toHaveBeenCalledTimes(2);',
      column: 28,
      endColumn: 40,
    },
    {
      code: 'expect((fn.mock).calls).toHaveLength(1);',
      output: 'expect((fn)).toHaveBeenCalledTimes(1);',
      column: 25,
      endColumn: 37,
    },
    {
      code: 'const assertion = expect(fn.mock.calls).toHaveLength(1);',
      output: null,
      column: 41,
      endColumn: 53,
    },
  ].map(({ column, endColumn, ...testCase }) => ({
    ...testCase,
    errors: [
      {
        messageId: 'preferMatcher',
        message: 'Prefer `toHaveBeenCalledTimes`',
        line: 1,
        column,
        endLine: 1,
        endColumn,
      },
    ],
  })),
});
