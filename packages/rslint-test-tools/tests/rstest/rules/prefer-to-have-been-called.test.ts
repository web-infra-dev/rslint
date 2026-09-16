import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-to-have-been-called', {} as never, {
  valid: [
    'expect(method.mock.calls).toHaveLength;',
    'expect(method).toBeCalledTimes();',
    'expect(method.mock.calls).toHaveLength(0);',
    'expect(method).toHaveBeenCalledTimes(1)',
    'expect(method).not.toHaveBeenCalledTimes(x)',
    'expect(method).not.toHaveBeenCalledTimes(1)',
    'expect(method).not.toHaveBeenCalledTimes(...x)',
    'expect(a);',
    'expect(method).toBe([])',
    'expect(fn.mock.calls).toEqual([])',
    'expect(fn.mock.calls).toContain(1, 2, 3)',
  ].map((code) => ({ code })),
  invalid: [
    {
      code: 'expect(method).toBeCalledTimes(0);',
      output: 'expect(method).not.toHaveBeenCalled();',
      column: 16,
      endColumn: 31,
    },
    {
      code: 'expect(method).not.toBeCalledTimes(0);',
      output: 'expect(method).toHaveBeenCalled();',
      column: 20,
      endColumn: 35,
    },
    {
      code: 'expect(method).toHaveBeenCalledTimes(0);',
      output: 'expect(method).not.toHaveBeenCalled();',
      column: 16,
      endColumn: 37,
    },
    {
      code: 'expect(method).not.toHaveBeenCalledTimes(0);',
      output: 'expect(method).toHaveBeenCalled();',
      column: 20,
      endColumn: 41,
    },
    {
      code: 'expect(method).not.toHaveBeenCalledTimes(0, 1, 2);',
      output: null,
      column: 20,
      endColumn: 41,
    },
    {
      code: 'expect(method).resolves.toHaveBeenCalledTimes(0);',
      output: 'expect(method).resolves.not.toHaveBeenCalled();',
      column: 25,
      endColumn: 46,
    },
    {
      code: 'expect(method).rejects.not.toHaveBeenCalledTimes(0);',
      output: 'expect(method).rejects.toHaveBeenCalled();',
      column: 28,
      endColumn: 49,
    },
    {
      code: 'expect(method).toBeCalledTimes(0 as number);',
      output: 'expect(method).not.toHaveBeenCalled();',
      column: 16,
      endColumn: 31,
    },
    {
      code: 'expect(method).not.resolves.toHaveBeenCalledTimes(0);',
      output: 'expect(method).resolves.toHaveBeenCalled();',
      column: 29,
      endColumn: 50,
    },
  ].map(({ column, endColumn, ...testCase }) => ({
    ...testCase,
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
  })),
});
