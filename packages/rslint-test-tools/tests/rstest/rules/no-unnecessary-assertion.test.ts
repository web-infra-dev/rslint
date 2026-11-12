import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-unnecessary-assertion', {
  valid: [
    { code: 'declare const value: string | null; expect(value).toBeNull();' },
    {
      code: 'declare const value: string | undefined; expect(value).to.be.undefined;',
    },
    {
      code: 'declare const value: Promise<string>; expect(value).resolves.toBeNull();',
    },
    { code: 'expect.poll((): string | null => null).toBeNull();' },
    { code: 'function check<T>(value: T) { expect(value).toBeNull(); }' },
    {
      code: 'declare function cleanup(): void; expect(cleanup()).toBeUndefined();',
    },
    {
      code: 'type UserId = number & { readonly __brand: unique symbol }; declare const value: UserId; expect(value).toBeNaN();',
    },
    { code: "expect.poll((): string => 'ready').to.be.null;" },
    { code: 'expect.element(locator).toBeNull();' },
  ],
  invalid: [
    {
      code: "expect('ready').toBeNull();",
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: "expect.soft('ready').to.not.be.undefined;",
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be undefined',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: "import { expect as check } from '@rstest/core'; check('ready').to.be.NaN;",
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be a number',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 73,
        },
      ],
    },
    {
      code: "expect.poll(async (): Promise<string> => 'ready').toBeNull();",
      errors: [
        {
          messageId: 'unnecessaryAssertion',
          message: 'Unnecessary assertion, subject cannot be null',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
  ],
});
