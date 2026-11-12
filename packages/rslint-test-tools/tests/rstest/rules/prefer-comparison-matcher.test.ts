import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-comparison-matcher', {
  valid: [
    'expect.hasAssertions',
    'expect.hasAssertions()',
    'expect.assertions(1)',
    'expect(true).toBe(...true)',
    'expect()',
    'expect({}).toStrictEqual({})',
    'expect(a).to.be.a("string");',
    'expect(a === b).toBe(true)',
    'expect(a !== 2).toStrictEqual(true)',
    'expect(a === b).not.toEqual(true)',
    'expect(a !== "string").toStrictEqual(true)',
    'expect(5 != a).toBe(true)',
    'expect(a == "string").toBe(true)',
    'expect(a == "string").not.toBe(true)',
    "expect().fail('Should not succeed a HTTPS proxy request.');",
  ].map((code) => ({ code })),
  invalid: [
    {
      code: 'expect(a > b).toBe(true)',
      errors: [
        {
          messageId: 'useToBeComparison',
          message: 'Prefer using `toBeGreaterThan` instead',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'expect(a > b).not.toBe(true)',
      errors: [
        {
          messageId: 'useToBeComparison',
          message: 'Prefer using `not.toBeGreaterThan` instead',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'expect(a < b).toBe(true)',
      errors: [
        {
          messageId: 'useToBeComparison',
          message: 'Prefer using `toBeLessThan` instead',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'expect(a < b).not.toBe(true)',
      errors: [
        {
          messageId: 'useToBeComparison',
          message: 'Prefer using `not.toBeLessThan` instead',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'expect(a >= b).toBe(true)',
      errors: [
        {
          messageId: 'useToBeComparison',
          message: 'Prefer using `toBeGreaterThanOrEqual` instead',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'expect(a >= b).not.toBe(true)',
      errors: [
        {
          messageId: 'useToBeComparison',
          message: 'Prefer using `not.toBeGreaterThanOrEqual` instead',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'expect(a <= b).toBe(true)',
      errors: [
        {
          messageId: 'useToBeComparison',
          message: 'Prefer using `toBeLessThanOrEqual` instead',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'expect(2 > 1).toBe(true)',
      errors: [
        {
          messageId: 'useToBeComparison',
          message: 'Prefer using `toBeGreaterThan` instead',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestComparisonMatcher',
              desc: 'Use `toBeGreaterThan`',
              output: 'expect(2).toBeGreaterThan(1)',
            },
          ],
        },
      ],
    },
  ],
});
