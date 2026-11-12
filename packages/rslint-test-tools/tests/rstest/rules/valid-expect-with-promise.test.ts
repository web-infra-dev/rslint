import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('valid-expect-with-promise', {
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
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: `(expect(Promise.resolve()) satisfies unknown).toBe(1)`,
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: `expect(Promise.resolve())!.toBe(1)`,
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: `expect({ value: Promise.resolve(1) }).property('value').toBe(1)`,
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 64,
        },
      ],
    },
    {
      code: `(import.meta.rstest.expect(1) as any).resolves.toBe(1)`,
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message: 'Subject is not a promise so resolves is not needed',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: 'expect("hello world").resolves.toContain(1)',
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message: 'Subject is not a promise so resolves is not needed',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'expect({}).rejects.not.toContain(1)',
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message: 'Subject is not a promise so rejects is not needed',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'declare const value: PromiseLike<number>; expect(value).toBe(1)',
      options: [{ checkThenables: true }],
      errors: [
        {
          messageId: 'poorlyExpectedPromise',
          message: 'Subject is a promise so resolve or reject should be used',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 64,
        },
      ],
    },
    {
      code: `class PromiseWithValue extends Promise<number> { value = 1 } declare const subject: PromiseWithValue; expect(subject).to.have.property('value').and.resolves.toBe(1)`,
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message: 'Subject is not a promise so resolves is not needed',
          line: 1,
          column: 149,
          endLine: 1,
          endColumn: 157,
        },
      ],
    },
    {
      code: `class PromiseWithValue extends Promise<number> { value = 1 } declare const subject: PromiseWithValue; (expect(subject) as any).to.have.property('value').and.resolves.toBe(1)`,
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message: 'Subject is not a promise so resolves is not needed',
          line: 1,
          column: 158,
          endLine: 1,
          endColumn: 166,
        },
      ],
    },
    {
      code: `function required(value: string): Promise<never> { return Promise.reject(new Error(value)); } expect(required).rejects.toThrow()`,
      errors: [
        {
          messageId: 'unneededRejectResolve',
          message: 'Subject is not a promise so rejects is not needed',
          line: 1,
          column: 112,
          endLine: 1,
          endColumn: 119,
        },
      ],
    },
  ],
});
