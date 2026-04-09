import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-alias-methods', {} as never, {
  valid: [
    { code: 'expect(a).toHaveBeenCalled()' },
    { code: 'expect(a).toHaveBeenCalledTimes()' },
    { code: 'expect(a).toHaveBeenCalledWith()' },
    { code: 'expect(a).toHaveBeenLastCalledWith()' },
    { code: 'expect(a).toHaveBeenNthCalledWith()' },
    { code: 'expect(a).toHaveReturned()' },
    { code: 'expect(a).toHaveReturnedTimes()' },
    { code: 'expect(a).toHaveReturnedWith()' },
    { code: 'expect(a).toHaveLastReturnedWith()' },
    { code: 'expect(a).toHaveNthReturnedWith()' },
    { code: 'expect(a).toThrow()' },
    { code: 'expect(a).rejects;' },
    { code: 'expect(a);' },
  ],

  invalid: [
    {
      code: 'expect(a).toBeCalled()',
      output: 'expect(a).toHaveBeenCalled()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toBeCalled',
            canonical: 'toHaveBeenCalled',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).toBeCalledTimes()',
      output: 'expect(a).toHaveBeenCalledTimes()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toBeCalledTimes',
            canonical: 'toHaveBeenCalledTimes',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).toBeCalledWith()',
      output: 'expect(a).toHaveBeenCalledWith()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toBeCalledWith',
            canonical: 'toHaveBeenCalledWith',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).lastCalledWith()',
      output: 'expect(a).toHaveBeenLastCalledWith()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'lastCalledWith',
            canonical: 'toHaveBeenLastCalledWith',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).nthCalledWith()',
      output: 'expect(a).toHaveBeenNthCalledWith()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'nthCalledWith',
            canonical: 'toHaveBeenNthCalledWith',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).toReturn()',
      output: 'expect(a).toHaveReturned()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toReturn',
            canonical: 'toHaveReturned',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).toReturnTimes()',
      output: 'expect(a).toHaveReturnedTimes()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toReturnTimes',
            canonical: 'toHaveReturnedTimes',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).toReturnWith()',
      output: 'expect(a).toHaveReturnedWith()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toReturnWith',
            canonical: 'toHaveReturnedWith',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).lastReturnedWith()',
      output: 'expect(a).toHaveLastReturnedWith()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'lastReturnedWith',
            canonical: 'toHaveLastReturnedWith',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).nthReturnedWith()',
      output: 'expect(a).toHaveNthReturnedWith()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'nthReturnedWith',
            canonical: 'toHaveNthReturnedWith',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).toThrowError()',
      output: 'expect(a).toThrow()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toThrowError',
            canonical: 'toThrow',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).resolves.toThrowError()',
      output: 'expect(a).resolves.toThrow()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toThrowError',
            canonical: 'toThrow',
          },
          column: 20,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).rejects.toThrowError()',
      output: 'expect(a).rejects.toThrow()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toThrowError',
            canonical: 'toThrow',
          },
          column: 19,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).not.toThrowError()',
      output: 'expect(a).not.toThrow()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toThrowError',
            canonical: 'toThrow',
          },
          column: 15,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).not["toThrowError"]()',
      output: 'expect(a).not["toThrow"]()',
      errors: [
        {
          messageId: 'replaceAlias',
          data: {
            alias: 'toThrowError',
            canonical: 'toThrow',
          },
          column: 15,
          line: 1,
        },
      ],
    },
  ],
});
