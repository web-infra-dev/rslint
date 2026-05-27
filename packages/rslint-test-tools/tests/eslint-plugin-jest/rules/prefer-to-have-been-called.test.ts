import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-to-have-been-called', {} as never, {
  valid: [
    { code: 'expect(method.mock.calls).toHaveLength;' },
    { code: 'expect(method.mock.calls).toHaveLength(0);' },
    { code: 'expect(method).toHaveBeenCalledTimes(1)' },
    { code: 'expect(method).not.toHaveBeenCalledTimes(x)' },
    { code: 'expect(method).not.toHaveBeenCalledTimes(1)' },
    { code: 'expect(method).not.toHaveBeenCalledTimes(...x)' },
    { code: 'expect(a);' },
    { code: 'expect(method).not.resolves.toHaveBeenCalledTimes(0);' },
    { code: 'expect(method).toBeCalledTimes(0!);' },
    { code: 'expect(method).toBeCalledTimes(0 satisfies number);' },
    { code: 'expect(method).toBe([])' },
    { code: 'expect(fn.mock.calls).toEqual([])' },
    { code: 'expect(fn.mock.calls).toContain(1, 2, 3)' },
  ],
  invalid: [
    {
      code: 'expect(method).toBeCalledTimes(0);',
      output: 'expect(method).not.toHaveBeenCalled();',
      errors: [{ messageId: 'preferMatcher', column: 16, line: 1 }],
    },
    {
      code: 'expect(method).not.toBeCalledTimes(0);',
      output: 'expect(method).toHaveBeenCalled();',
      errors: [{ messageId: 'preferMatcher', column: 20, line: 1 }],
    },
    {
      code: 'expect(method).toHaveBeenCalledTimes(0);',
      output: 'expect(method).not.toHaveBeenCalled();',
      errors: [{ messageId: 'preferMatcher', column: 16, line: 1 }],
    },
    {
      code: 'expect(method).not.toHaveBeenCalledTimes(0);',
      output: 'expect(method).toHaveBeenCalled();',
      errors: [{ messageId: 'preferMatcher', column: 20, line: 1 }],
    },
    {
      code: 'expect(method).not.toHaveBeenCalledTimes(0, 1, 2);',
      output: 'expect(method).toHaveBeenCalled();',
      errors: [{ messageId: 'preferMatcher', column: 20, line: 1 }],
    },

    {
      code: 'expect(method).resolves.toHaveBeenCalledTimes(0);',
      output: 'expect(method).resolves.not.toHaveBeenCalled();',
      errors: [{ messageId: 'preferMatcher', column: 25, line: 1 }],
    },
    {
      code: 'expect(method).rejects.not.toHaveBeenCalledTimes(0);',
      output: 'expect(method).rejects.toHaveBeenCalled();',
      errors: [{ messageId: 'preferMatcher', column: 28, line: 1 }],
    },

    {
      code: 'expect(method).toBeCalledTimes(0 as number);',
      output: 'expect(method).not.toHaveBeenCalled();',
      errors: [{ messageId: 'preferMatcher', column: 16, line: 1 }],
    },
  ],
});
