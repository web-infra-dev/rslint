import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unnecessary-assertion', {} as never, {
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
          message: 'Unnecessary assertion, subject cannot be null',
        },
      ],
    },
    {
      code: "expect.soft('ready').to.not.be.undefined;",
      errors: [
        {
          message: 'Unnecessary assertion, subject cannot be undefined',
        },
      ],
    },
    {
      code: "import { expect as check } from '@rstest/core'; check('ready').to.be.NaN;",
      errors: [
        {
          message: 'Unnecessary assertion, subject cannot be a number',
        },
      ],
    },
    {
      code: "expect.poll(async (): Promise<string> => 'ready').toBeNull();",
      errors: [
        {
          message: 'Unnecessary assertion, subject cannot be null',
        },
      ],
    },
  ],
});
