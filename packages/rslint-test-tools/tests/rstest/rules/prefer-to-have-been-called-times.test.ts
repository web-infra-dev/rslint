import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-to-have-been-called-times', {
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
    'expect(fn.mock.calls).property(0).toHaveLength(1);',
    'expect.poll(fn.mock.calls).toHaveLength(1);',
    'expect.element(fn.mock.calls).toHaveLength(1);',
  ].map((code) => ({ code })),
  invalid: [
    {
      code: 'expect(method.mock.calls).toHaveLength(1);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 39,
        },
      ],
      output: 'expect(method).toHaveBeenCalledTimes(1);',
    },
    {
      code: 'expect(method.mock.calls).resolves.toHaveLength(x);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: "expect(fn.mock.calls).toHaveLength('1');",
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'expect(fn.mock.calls).toHaveLength((fn.mockClear(), 1));',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'class Derived extends Base { static verify() { expect(super.mock.calls).toHaveLength(1); } }',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 73,
          endLine: 1,
          endColumn: 85,
        },
      ],
    },
    {
      code: 'expect(method["mock"].calls).toHaveLength(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 42,
        },
      ],
      output: 'expect(method).toHaveBeenCalledTimes(0);',
    },
    {
      code: 'expect(my.method.mock.calls).not.toHaveLength(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 46,
        },
      ],
      output: 'expect(my.method).not.toHaveBeenCalledTimes(0);',
    },
    {
      code: 'expect<unknown[]>(fn.mock.calls).toHaveLength(0);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 46,
        },
      ],
      output: 'expect(fn).toHaveBeenCalledTimes(0);',
    },
    {
      code: 'expect.soft(fn.mock.calls).toHaveLength(2);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 40,
        },
      ],
      output: 'expect.soft(fn).toHaveBeenCalledTimes(2);',
    },
    {
      code: 'expect((fn.mock).calls).toHaveLength(1);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: 'expect((fn)).toHaveBeenCalledTimes(1);',
    },
    {
      code: 'const assertion = expect(fn.mock.calls).toHaveLength(1);',
      errors: [
        {
          messageId: 'preferMatcher',
          message: 'Prefer `toHaveBeenCalledTimes`',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
  ],
});
