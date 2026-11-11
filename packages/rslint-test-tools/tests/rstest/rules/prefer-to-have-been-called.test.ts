import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-to-have-been-called', {
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
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'expect(method).not.toHaveBeenCalled();',
    },
    {
      code: 'expect(method).not.toBeCalledTimes(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 35,
        },
      ],
      output: 'expect(method).toHaveBeenCalled();',
    },
    {
      code: 'expect(method).toHaveBeenCalledTimes(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: 'expect(method).not.toHaveBeenCalled();',
    },
    {
      code: 'expect(method).not.toHaveBeenCalledTimes(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 41,
        },
      ],
      output: 'expect(method).toHaveBeenCalled();',
    },
    {
      code: 'expect(method).not.toHaveBeenCalledTimes(0, 1, 2);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'expect(method).resolves.toHaveBeenCalledTimes(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 46,
        },
      ],
      output: 'expect(method).resolves.not.toHaveBeenCalled();',
    },
    {
      code: 'expect(method).rejects.not.toHaveBeenCalledTimes(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 49,
        },
      ],
      output: 'expect(method).rejects.toHaveBeenCalled();',
    },
    {
      code: 'expect(method).toBeCalledTimes(0 as number);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'expect(method).not.toHaveBeenCalled();',
    },
    {
      code: 'expect(method).not.resolves.toHaveBeenCalledTimes(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Use `toHaveBeenCalled`',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 50,
        },
      ],
      output: 'expect(method).resolves.toHaveBeenCalled();',
    },
  ],
});
