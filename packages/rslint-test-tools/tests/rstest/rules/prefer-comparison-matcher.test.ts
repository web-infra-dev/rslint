import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-comparison-matcher', {} as never, {
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
    ...[
      ['>', 'toBeGreaterThan'],
      ['<', 'toBeLessThan'],
      ['>=', 'toBeGreaterThanOrEqual'],
      ['<=', 'toBeLessThanOrEqual'],
    ].flatMap(([operator, matcher]) =>
      [false, true]
        .filter((negated) => !negated || operator !== '<=')
        .map((negated) => {
          const modifier = negated ? 'not.' : '';
          return {
            code: `expect(a ${operator} b).${modifier}toBe(true)`,
            errors: [
              {
                messageId: 'useToBeComparison',
                message: `Prefer using \`${modifier}${matcher}\` instead`,
              },
            ],
          };
        }),
    ),
    {
      code: 'expect(2 > 1).toBe(true)',
      errors: [
        {
          messageId: 'useToBeComparison',
          suggestions: [
            {
              messageId: 'suggestComparisonMatcher',
              output: 'expect(2).toBeGreaterThan(1)',
            },
          ],
        },
      ],
    },
  ],
});
