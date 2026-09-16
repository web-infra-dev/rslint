import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-expect-with-promise', {} as never, {
  valid: [
    { code: 'expect.hasAssertions()' },
    { code: 'expect(1).toBe(1)' },
    { code: 'expect(Promise.resolve()).resolves.toBe(1)' },
    { code: 'expect(Promise.resolve()).rejects.toBe(1)' },
    {
      code: 'declare const value: PromiseLike<number>; expect(value).resolves.toBe(1)',
      options: [{ checkThenables: true }],
    },
    {
      code: `declare const subject: { value: Promise<number> }; expect(subject).to.have.property('value').and.resolves.toBe(1)`,
    },
    {
      code: `declare const subject: { value: Promise<number> }; (expect(subject) as any).to.have.property('value').and.resolves.toBe(1)`,
    },
    { code: `expect({ value: Promise.resolve(1) }).property('value')` },
    {
      code: `expect({ value: Promise.resolve(1) }).property('value').resolves.toBe(1)`,
    },
  ],
  invalid: [
    {
      code: 'expect(Promise.resolve()).toBe(1)',
      errors: [
        {
          message: 'Subject is a promise so resolve or reject should be used',
          messageId: 'poorlyExpectedPromise',
          line: 1,
          column: 1,
        },
      ],
    },
    {
      code: `(expect(Promise.resolve()) satisfies unknown).toBe(1)`,
      errors: [{ messageId: 'poorlyExpectedPromise' }],
    },
    {
      code: `expect(Promise.resolve())!.toBe(1)`,
      errors: [{ messageId: 'poorlyExpectedPromise' }],
    },
    {
      code: `expect({ value: Promise.resolve(1) }).property('value').toBe(1)`,
      errors: [{ messageId: 'poorlyExpectedPromise' }],
    },
    {
      code: `(import.meta.rstest.expect(1) as any).resolves.toBe(1)`,
      errors: [{ messageId: 'unneededRejectResolve' }],
    },
    {
      code: 'expect("hello world").resolves.toContain(1)',
      errors: [
        {
          message: 'Subject is not a promise so resolves is not needed',
          messageId: 'unneededRejectResolve',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'expect({}).rejects.not.toContain(1)',
      errors: [
        {
          message: 'Subject is not a promise so rejects is not needed',
          messageId: 'unneededRejectResolve',
          line: 1,
          column: 12,
        },
      ],
    },
    {
      code: 'declare const value: PromiseLike<number>; expect(value).toBe(1)',
      options: [{ checkThenables: true }],
      errors: [{ messageId: 'poorlyExpectedPromise' }],
    },
    {
      code: `class PromiseWithValue extends Promise<number> { value = 1 } declare const subject: PromiseWithValue; expect(subject).to.have.property('value').and.resolves.toBe(1)`,
      errors: [{ messageId: 'unneededRejectResolve' }],
    },
    {
      code: `class PromiseWithValue extends Promise<number> { value = 1 } declare const subject: PromiseWithValue; (expect(subject) as any).to.have.property('value').and.resolves.toBe(1)`,
      errors: [{ messageId: 'unneededRejectResolve' }],
    },
    {
      code: `function required(value: string): Promise<never> { return Promise.reject(new Error(value)); } expect(required).rejects.toThrow()`,
      errors: [{ messageId: 'unneededRejectResolve' }],
    },
  ],
});
