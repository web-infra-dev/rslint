import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-alias-methods', {} as never, {
  valid: [
    { code: `expect(a).toHaveBeenCalled()` },
    { code: `expect(a).toHaveBeenCalledTimes()` },
    { code: `expect(a).toHaveBeenCalledWith()` },
    { code: `expect(a).toHaveBeenLastCalledWith()` },
    { code: `expect(a).toHaveBeenNthCalledWith()` },
    { code: `expect(a).toHaveReturned()` },
    { code: `expect(a).toHaveReturnedTimes()` },
    { code: `expect(a).toHaveReturnedWith()` },
    { code: `expect(a).toHaveLastReturnedWith()` },
    { code: `expect(a).toHaveNthReturnedWith()` },
    { code: `expect(a).toThrow()` },
    { code: `expect(a).rejects;` },
    { code: `expect(a);` },
    { code: `expect(a).to.be.a("string");` },
  ],
  invalid: [
    {
      code: `expect(a).toBeCalled()`,
      output: `expect(a).toHaveBeenCalled()`,
      errors: [
        {
          messageId: 'replaceAlias',
          message:
            'Replace toBeCalled() with its canonical name of toHaveBeenCalled()',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: `expect(a).toBeCalledTimes()`,
      output: `expect(a).toHaveBeenCalledTimes()`,
      errors: [
        {
          messageId: 'replaceAlias',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: `expect(a).not["toThrowError"]()`,
      output: `expect(a).not["toThrow"]()`,
      errors: [
        {
          messageId: 'replaceAlias',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
  ],
});
