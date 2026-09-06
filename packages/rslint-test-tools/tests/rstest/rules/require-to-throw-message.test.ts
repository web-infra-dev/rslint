import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-to-throw-message', {} as never, {
  valid: [
    { code: "expect(() => { throw new Error('a'); }).toThrow('a');" },
    { code: "expect(() => { throw new Error('a'); }).toThrowError('a');" },
    {
      code: `
        test('string', async () => {
          const throwErrorAsync = async () => { throw new Error('a') };
          await expect(throwErrorAsync()).rejects.toThrow('a');
          await expect(throwErrorAsync()).rejects.toThrowError('a');
        })
      `,
    },
    {
      code: "const a = 'a'; expect(() => { throw new Error('a'); }).toThrow(`${a}`);",
    },
    {
      code: "const a = 'a'; expect(() => { throw new Error('a'); }).toThrowError(`${a}`);",
    },
    {
      code:
        "test('Template literal', async () => {\n" +
        "  const a = 'a';\n" +
        "  const throwErrorAsync = async () => { throw new Error('a') };\n" +
        '  await expect(throwErrorAsync()).rejects.toThrow(`${a}`);\n' +
        '  await expect(throwErrorAsync()).rejects.toThrowError(`${a}`);\n' +
        '})',
    },
    { code: "expect(() => { throw new Error('a'); }).toThrow(/^a$/);" },
    {
      code: "expect(() => { throw new Error('a'); }).toThrowError(/^a$/);",
    },
    {
      code: `
        test('Regex', async () => {
          const throwErrorAsync = async () => { throw new Error('a') };
          await expect(throwErrorAsync()).rejects.toThrow(/^a$/);
          await expect(throwErrorAsync()).rejects.toThrowError(/^a$/);
        })
      `,
    },
    {
      code: "expect(() => { throw new Error('a'); }).toThrow((() => { return 'a'; })());",
    },
    {
      code: "expect(() => { throw new Error('a'); }).toThrowError((() => { return 'a'; })());",
    },
    {
      code: `
        test('Function', async () => {
          const throwErrorAsync = async () => { throw new Error('a') };
          const fn = () => { return 'a'; };
          await expect(throwErrorAsync()).rejects.toThrow(fn());
          await expect(throwErrorAsync()).rejects.toThrowError(fn());
        })
      `,
    },
    { code: "expect(() => { throw new Error('a'); }).not.toThrow();" },
    {
      code: "expect(() => { throw new Error('a'); }).not.toThrowError();",
    },
    {
      code: `
        test('Allow no message for not', async () => {
          const throwErrorAsync = async () => { throw new Error('a') };
          await expect(throwErrorAsync()).resolves.not.toThrow();
          await expect(throwErrorAsync()).resolves.not.toThrowError();
        })
      `,
    },
    { code: 'expect(a);' },
  ],
  invalid: [
    {
      code: "expect(() => { throw new Error('a'); }).toThrow();",
      errors: [
        {
          messageId: 'addErrorMessage',
          message: 'Add an error message to toThrow()',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: "expect(() => { throw new Error('a'); }).toThrowError();",
      errors: [
        {
          messageId: 'addErrorMessage',
          message: 'Add an error message to toThrowError()',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: `test('empty rejects.toThrow', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow();
  await expect(throwErrorAsync()).rejects.toThrowError();
})`,
      errors: [
        {
          messageId: 'addErrorMessage',
          message: 'Add an error message to toThrow()',
          line: 3,
          column: 43,
          endLine: 3,
          endColumn: 50,
        },
        {
          messageId: 'addErrorMessage',
          message: 'Add an error message to toThrowError()',
          line: 4,
          column: 43,
          endLine: 4,
          endColumn: 55,
        },
      ],
    },
  ],
});
