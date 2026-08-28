import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-called-once', {} as never, {
  valid: [
    { code: `expect(fn).toBeCalledOnce();` },
    { code: `expect(fn).toHaveBeenCalledOnce();` },
    { code: `expect(fn).toBeCalledTimes(2);` },
    { code: `expect(fn).toHaveBeenCalledTimes(2);` },
    { code: `expect(fn).toBeCalledTimes(expect.anything());` },
    { code: `expect(fn).toHaveBeenCalledTimes(expect.anything());` },
    { code: `expect(fn).not.toBeCalledOnce();` },
    { code: `expect(fn).rejects.not.toBeCalledOnce();` },
    { code: `expect(fn).not.toHaveBeenCalledOnce();` },
    { code: `expect(fn).resolves.not.toHaveBeenCalledOnce();` },
    { code: `expect(fn).toBeCalledTimes(0);` },
    { code: `expect(fn).toHaveBeenCalledTimes(0);` },
    { code: `expect(fn);` },
  ],
  invalid: [
    {
      code: `expect(fn).toBeCalledTimes(1);`,
      output: `expect(fn).toHaveBeenCalledOnce();`,
      errors: [
        {
          messageId: 'preferCalledOnce',
          message: 'Prefer toHaveBeenCalledOnce()',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: `expect(fn).toHaveBeenCalledTimes(1);`,
      output: `expect(fn).toHaveBeenCalledOnce();`,
      errors: [
        {
          messageId: 'preferCalledOnce',
          message: 'Prefer toHaveBeenCalledOnce()',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: `expect(fn).not.toBeCalledTimes(1);`,
      output: `expect(fn).not.toHaveBeenCalledOnce();`,
      errors: [
        {
          messageId: 'preferCalledOnce',
          message: 'Prefer toHaveBeenCalledOnce()',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: `expect(fn).not.toHaveBeenCalledTimes(1);`,
      output: `expect(fn).not.toHaveBeenCalledOnce();`,
      errors: [
        {
          messageId: 'preferCalledOnce',
          message: 'Prefer toHaveBeenCalledOnce()',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: `expect(fn).resolves.toBeCalledTimes(1);`,
      output: `expect(fn).resolves.toHaveBeenCalledOnce();`,
      errors: [
        {
          messageId: 'preferCalledOnce',
          message: 'Prefer toHaveBeenCalledOnce()',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: `expect(fn).resolves.toHaveBeenCalledTimes(1);`,
      output: `expect(fn).resolves.toHaveBeenCalledOnce();`,
      errors: [
        {
          messageId: 'preferCalledOnce',
          message: 'Prefer toHaveBeenCalledOnce()',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
  ],
});
