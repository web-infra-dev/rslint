import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-restricted-matchers', {} as never, {
  valid: [
    { code: 'expect(a).toHaveBeenCalled()' },
    { code: 'expect(a).not.toHaveBeenCalled()' },
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
    {
      code: 'expect(a).resolves',
      options: [{ not: null }],
    },
    {
      code: 'expect(a).toBe(b)',
      options: [{ 'not.toBe': null }],
    },
    {
      code: 'expect(a).toBeUndefined(b)',
      options: [{ toBe: null }],
    },
    {
      code: 'expect(a)["toBe"](b)',
      options: [{ 'not.toBe': null }],
    },
    {
      code: 'expect(a).resolves.not.toBe(b)',
      options: [{ not: null }],
    },
    {
      code: 'expect(a).resolves.not.toBe(b)',
      options: [{ 'not.toBe': null }],
    },
    {
      code: "expect(uploadFileMock).resolves.toHaveBeenCalledWith('file.name')",
      options: [
        { 'not.toHaveBeenCalledWith': 'Use not.toHaveBeenCalled instead' },
      ],
    },
    {
      code: "expect(uploadFileMock).resolves.not.toHaveBeenCalledWith('file.name')",
      options: [
        { 'not.toHaveBeenCalledWith': 'Use not.toHaveBeenCalled instead' },
      ],
    },
  ],
  invalid: [
    {
      code: 'expect(a).toBe(b)',
      options: [{ toBe: null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'toBe',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a)["toBe"](b)',
      options: [{ toBe: null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'toBe',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).not[x]()',
      options: [{ not: null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'not',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).not.toBe(b)',
      options: [{ not: null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'not',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).resolves.toBe(b)',
      options: [{ resolves: null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'resolves',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).resolves.not.toBe(b)',
      options: [{ resolves: null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'resolves',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).resolves.not.toBe(b)',
      options: [{ 'resolves.not': null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'resolves.not',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).not.toBe(b)',
      options: [{ 'not.toBe': null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'not.toBe',
          },
          endColumn: 19,
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).resolves.not.toBe(b)',
      options: [{ 'resolves.not.toBe': null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'resolves.not.toBe',
          },
          endColumn: 28,
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(a).toBe(b)',
      options: [{ toBe: 'Prefer `toStrictEqual` instead' }],
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          data: {
            message: 'Prefer `toStrictEqual` instead',
            restriction: 'toBe',
          },
          column: 11,
          line: 1,
        },
      ],
    },
    {
      code: `
        test('some test', async () => {
          await expect(Promise.resolve(1)).resolves.toBe(1);
         });
      `,
      options: [{ resolves: 'Use `expect(await promise)` instead.' }],
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          data: {
            message: 'Use `expect(await promise)` instead.',
            restriction: 'resolves',
          },
          endColumn: 57,
          column: 44,
        },
      ],
    },
    {
      code: 'expect(Promise.resolve({})).rejects.toBeFalsy()',
      options: [{ 'rejects.toBeFalsy': null }],
      errors: [
        {
          messageId: 'restrictedChain',
          data: {
            message: null,
            restriction: 'rejects.toBeFalsy',
          },
          endColumn: 46,
          column: 29,
        },
      ],
    },
    {
      code: "expect(uploadFileMock).not.toHaveBeenCalledWith('file.name')",
      options: [
        { 'not.toHaveBeenCalledWith': 'Use not.toHaveBeenCalled instead' },
      ],
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          data: {
            message: 'Use not.toHaveBeenCalled instead',
            restriction: 'not.toHaveBeenCalledWith',
          },
          endColumn: 48,
          column: 24,
        },
      ],
    },
  ],
});
