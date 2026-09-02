import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-strict-boolean-matchers', {} as never, {
  valid: [
    { code: `[].push(true)` },
    { code: `[].push(false)` },
    { code: `expect("something");` },
    { code: `expect(true).toBe(true);` },
    { code: `expect(false).toBe(false);` },
    { code: `expect(value).toEqual();` },
    { code: `expect(value).not.toBe(true);` },
    { code: `expect(value).toBe(undefined);` },
    { code: `expect(value).toBe();` },
    { code: `expect(true).toMatchSnapshot();` },
    { code: `expect("a string").toMatchSnapshot(true);` },
    { code: `expect(something).toEqual('a string');` },
    { code: `expect(true).toBe` },
    { code: `expect(value).to.be.a("boolean");` },
  ],
  invalid: [
    {
      code: `expect(false).toBeTruthy();`,
      output: `expect(false).toBe(true);`,
      errors: [
        {
          messageId: 'preferToBeTrue',
          message: 'Prefer using `toBe(true)` to test value is `true`',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: `expect(false).toBeFalsy();`,
      output: `expect(false).toBe(false);`,
      errors: [
        {
          messageId: 'preferToBeFalse',
          message: 'Prefer using `toBe(false)` to test value is `false`',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: `expect(wasSuccessful).toBeTruthy();`,
      output: `expect(wasSuccessful).toBe(true);`,
      errors: [
        {
          messageId: 'preferToBeTrue',
          message: 'Prefer using `toBe(true)` to test value is `true`',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: `expect("a string").not.toBeTruthy();`,
      output: `expect("a string").not.toBe(true);`,
      errors: [
        {
          messageId: 'preferToBeTrue',
          message: 'Prefer using `toBe(true)` to test value is `true`',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: `expect("a string").not.toBeFalsy();`,
      output: `expect("a string").not.toBe(false);`,
      errors: [
        {
          messageId: 'preferToBeFalse',
          message: 'Prefer using `toBe(false)` to test value is `false`',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
  ],
});
