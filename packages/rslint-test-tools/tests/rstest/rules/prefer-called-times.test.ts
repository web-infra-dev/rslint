import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-called-times', {} as never, {
  valid: [
    { code: `expect(fn).toBeCalledTimes(1);` },
    { code: `expect(fn).toHaveBeenCalledTimes(1);` },
    { code: `expect(fn).toBeCalledTimes(2);` },
    { code: `expect(fn).toHaveBeenCalledTimes(2);` },
    { code: `expect(fn).toBeCalledTimes(expect.anything());` },
    { code: `expect(fn).toHaveBeenCalledTimes(expect.anything());` },
    { code: `expect(fn).not.toBeCalledTimes(2);` },
    { code: `expect(fn).rejects.not.toBeCalledTimes(1);` },
    { code: `expect(fn).not.toHaveBeenCalledTimes(1);` },
    { code: `expect(fn).resolves.not.toHaveBeenCalledTimes(1);` },
    { code: `expect(fn).toBeCalledTimes(0);` },
    { code: `expect(fn).toHaveBeenCalledTimes(0);` },
    { code: `expect(fn);` },
    { code: `expect(fn).toBeCalledOnce();` },
    { code: `expect(fn).not.toBeCalledOnce();` },
    { code: `expect(fn).resolves.toBeCalledOnce();` },
  ],
  invalid: [
    {
      code: `expect(fn).toHaveBeenCalledOnce();`,
      output: `expect(fn).toHaveBeenCalledTimes(1);`,
      errors: [
        {
          messageId: 'preferCalledTimes',
          message: 'Prefer toHaveBeenCalledTimes(1)',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: `expect(fn).not.toHaveBeenCalledOnce();`,
      output: `expect(fn).not.toHaveBeenCalledTimes(1);`,
      errors: [
        {
          messageId: 'preferCalledTimes',
          message: 'Prefer toHaveBeenCalledTimes(1)',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: `expect(fn).resolves.toHaveBeenCalledOnce();`,
      output: `expect(fn).resolves.toHaveBeenCalledTimes(1);`,
      errors: [
        {
          messageId: 'preferCalledTimes',
          message: 'Prefer toHaveBeenCalledTimes(1)',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
  ],
});
