import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-to-be-falsy', {} as never, {
  valid: [
    { code: `[].push(false)` },
    { code: `expect("something");` },
    { code: `expect(value).to.be.a("boolean");` },
    { code: `expect(true).toBeTrue();` },
    { code: `expect(value).toEqual();` },
    { code: `expect(value).not.toBeTrue();` },
    { code: `expect(value).toBe(undefined);` },
    { code: `expect(false).toBe(true)` },
    { code: `expect(value).toBe();` },
    { code: `expect(false).toMatchSnapshot();` },
    { code: `expect("a string").toMatchSnapshot(false);` },
    { code: `expect(something).toEqual('a string');` },
    { code: `expect(false).toBe` },
  ],
  invalid: [
    {
      code: `expect(true).toBe(false);`,
      output: `expect(true).toBeFalsy();`,
      errors: [
        {
          messageId: 'preferToBeFalsy',
          message: 'Prefer using toBeFalsy()',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: `expect(wasSuccessful).toEqual(false);`,
      output: `expect(wasSuccessful).toBeFalsy();`,
      errors: [
        {
          messageId: 'preferToBeFalsy',
          message: 'Prefer using toBeFalsy()',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: `expect(fs.existsSync('/path/to/file')).toStrictEqual(false);`,
      output: `expect(fs.existsSync('/path/to/file')).toBeFalsy();`,
      errors: [
        {
          messageId: 'preferToBeFalsy',
          message: 'Prefer using toBeFalsy()',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: `expect("a string").not.toBe(false);`,
      output: `expect("a string").not.toBeFalsy();`,
      errors: [
        {
          messageId: 'preferToBeFalsy',
          message: 'Prefer using toBeFalsy()',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: `expect("a string").not.toEqual(false);`,
      output: `expect("a string").not.toBeFalsy();`,
      errors: [
        {
          messageId: 'preferToBeFalsy',
          message: 'Prefer using toBeFalsy()',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
  ],
});
