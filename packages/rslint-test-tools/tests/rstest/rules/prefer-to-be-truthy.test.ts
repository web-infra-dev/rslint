import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-to-be-truthy', {} as never, {
  valid: [
    { code: `[].push(true)` },
    { code: `expect("something");` },
    { code: `expect(value).to.be.a("boolean");` },
    { code: `expect(true).toBeTrue();` },
    { code: `expect(value).toEqual();` },
    { code: `expect(value).not.toBeTrue();` },
    { code: `expect(value).toBe(undefined);` },
    { code: `expect(true).toBe(false)` },
    { code: `expect(value).toBe();` },
    { code: `expect(true).toMatchSnapshot();` },
    { code: `expect("a string").toMatchSnapshot(true);` },
    { code: `expect(something).toEqual('a string');` },
    { code: `expect(true).toBe` },
  ],
  invalid: [
    {
      code: `expect(false).toBe(true);`,
      output: `expect(false).toBeTruthy();`,
      errors: [
        {
          messageId: 'preferToBeTruthy',
          message: 'Prefer using `toBeTruthy` to test value is `true`',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: `expect(wasSuccessful).toEqual(true);`,
      output: `expect(wasSuccessful).toBeTruthy();`,
      errors: [
        {
          messageId: 'preferToBeTruthy',
          message: 'Prefer using `toBeTruthy` to test value is `true`',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: `expect(fs.existsSync('/path/to/file')).toStrictEqual(true);`,
      output: `expect(fs.existsSync('/path/to/file')).toBeTruthy();`,
      errors: [
        {
          messageId: 'preferToBeTruthy',
          message: 'Prefer using `toBeTruthy` to test value is `true`',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: `expect("a string").not.toBe(true);`,
      output: `expect("a string").not.toBeTruthy();`,
      errors: [
        {
          messageId: 'preferToBeTruthy',
          message: 'Prefer using `toBeTruthy` to test value is `true`',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: `expect("a string").not.toEqual(true);`,
      output: `expect("a string").not.toBeTruthy();`,
      errors: [
        {
          messageId: 'preferToBeTruthy',
          message: 'Prefer using `toBeTruthy` to test value is `true`',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
  ],
});
