import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-restricted-matchers', {} as never, {
  valid: [
    { code: `expect(a).toHaveBeenCalled()` },
    { code: `expect(a).not.toHaveBeenCalled()` },
    { code: `expect(a).rejects;` },
    { code: `expect(a);` },
    { code: `expect(a).resolves`, options: [{ not: null }] },
    { code: `expect(a).toBe(b)`, options: [{ 'not.toBe': null }] },
    { code: `expect(a).toBeUndefined(b)`, options: [{ toBe: null }] },
    {
      code: `expect(a).to.be.a('string').and.contain('x')`,
      options: [{ contain: null }],
    },
  ],
  invalid: [
    {
      code: `expect(a).not.toBe(b)`,
      options: [{ not: null }],
      errors: [
        {
          messageId: 'restrictedChain',
          message: 'Use of `not` is disallowed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: `expect(a).to.be.a('string').and.contain('x')`,
      options: [{ 'to.be.a': null }],
      errors: [
        {
          messageId: 'restrictedChain',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: `expect(a).resolves.not.toBe(b)`,
      options: [{ 'resolves.not': null }],
      errors: [
        {
          messageId: 'restrictedChain',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: `expect(a).toBe(b)`,
      options: [{ toBe: 'Prefer `toStrictEqual` instead' }],
      errors: [
        {
          messageId: 'restrictedChainWithMessage',
          message: 'Prefer `toStrictEqual` instead',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
  ],
});
