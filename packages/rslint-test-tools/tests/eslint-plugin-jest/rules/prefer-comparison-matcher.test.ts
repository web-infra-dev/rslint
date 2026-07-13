import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const generateInvalidCases = (
  operator: string,
  equalityMatcher: string,
  preferredMatcher: string,
  preferredMatcherWhenNegated: string,
) => [
  {
    code: `expect(value ${operator} 1).${equalityMatcher}(true);`,
    output: `expect(value).${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher },
        column: 18 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1,).${equalityMatcher}(true,);`,
    output: `expect(value,).${preferredMatcher}(1,);`,
    parserOptions: { ecmaVersion: 2017 },
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher },
        column: 19 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)['${equalityMatcher}'](true);`,
    output: `expect(value).${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher },
        column: 18 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).resolves.${equalityMatcher}(true);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher },
        column: 27 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).${equalityMatcher}(false);`,
    output: `expect(value).${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher: preferredMatcherWhenNegated },
        column: 18 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)['${equalityMatcher}'](false);`,
    output: `expect(value).${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher: preferredMatcherWhenNegated },
        column: 18 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).resolves.${equalityMatcher}(false);`,
    output: `expect(value).resolves.${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher: preferredMatcherWhenNegated },
        column: 27 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).not.${equalityMatcher}(true);`,
    output: `expect(value).${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher: preferredMatcherWhenNegated },
        column: 22 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)['not'].${equalityMatcher}(true);`,
    output: `expect(value).${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher: preferredMatcherWhenNegated },
        column: 25 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).resolves.not.${equalityMatcher}(true);`,
    output: `expect(value).resolves.${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher: preferredMatcherWhenNegated },
        column: 31 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).not.${equalityMatcher}(false);`,
    output: `expect(value).${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher },
        column: 22 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).resolves.not.${equalityMatcher}(false);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher },
        column: 31 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)["resolves"].not.${equalityMatcher}(false);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher },
        column: 34 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)["resolves"]["not"].${equalityMatcher}(false);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher },
        column: 37 + operator.length,
        line: 1,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)["resolves"]["not"]['${equalityMatcher}'](false);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        data: { preferredMatcher },
        column: 37 + operator.length,
        line: 1,
      },
    ],
  },
];

const generateValidStringLiteralCases = (operator: string, matcher: string) =>
  [
    ['x', "'y'"],
    ['x', '`y`'],
    ['x', '`y${z}`'],
  ].flatMap(([a, b]) => [
    `expect(${a} ${operator} ${b}).${matcher}(true)`,
    `expect(${a} ${operator} ${b}).${matcher}(false)`,
    `expect(${a} ${operator} ${b}).not.${matcher}(true)`,
    `expect(${a} ${operator} ${b}).not.${matcher}(false)`,
    `expect(${a} ${operator} ${b}).resolves.${matcher}(true)`,
    `expect(${a} ${operator} ${b}).resolves.${matcher}(false)`,
    `expect(${a} ${operator} ${b}).resolves.not.${matcher}(true)`,
    `expect(${a} ${operator} ${b}).resolves.not.${matcher}(false)`,
    `expect(${b} ${operator} ${a}).resolves.not.${matcher}(false)`,
    `expect(${b} ${operator} ${a}).resolves.not.${matcher}(true)`,
    `expect(${b} ${operator} ${a}).resolves.${matcher}(false)`,
    `expect(${b} ${operator} ${a}).resolves.${matcher}(true)`,
    `expect(${b} ${operator} ${a}).not.${matcher}(false)`,
    `expect(${b} ${operator} ${a}).not.${matcher}(true)`,
    `expect(${b} ${operator} ${a}).${matcher}(false)`,
    `expect(${b} ${operator} ${a}).${matcher}(true)`,
  ]);

const testComparisonOperator = (
  operator: string,
  preferredMatcher: string,
  preferredMatcherWhenNegated: string,
) => {
  ruleTester.run(`prefer-comparison-matcher`, {} as never, {
    valid: [
      { code: 'expect()' },
      { code: 'expect({}).toStrictEqual({})' },
      { code: `expect(value).${preferredMatcher}(1);` },
      { code: `expect(value).${preferredMatcherWhenNegated}(1);` },
      { code: `expect(value).not.${preferredMatcher}(1);` },
      { code: `expect(value).not.${preferredMatcherWhenNegated}(1);` },
      ...['toBe', 'toEqual', 'toStrictEqual'].flatMap(
        (equalityMatcher: string) =>
          generateValidStringLiteralCases(operator, equalityMatcher).map(
            (code) => ({ code }),
          ),
      ),
    ],
    invalid: ['toBe', 'toEqual', 'toStrictEqual'].flatMap((equalityMatcher) =>
      generateInvalidCases(
        operator,
        equalityMatcher,
        preferredMatcher,
        preferredMatcherWhenNegated,
      ),
    ),
  });
};

testComparisonOperator('>', 'toBeGreaterThan', 'toBeLessThanOrEqual');
testComparisonOperator('<', 'toBeLessThan', 'toBeGreaterThanOrEqual');
testComparisonOperator('>=', 'toBeGreaterThanOrEqual', 'toBeLessThan');
testComparisonOperator('<=', 'toBeLessThanOrEqual', 'toBeGreaterThan');

ruleTester.run(`prefer-comparison-matcher`, {} as never, {
  valid: [
    { code: 'expect.hasAssertions' },
    { code: 'expect.hasAssertions()' },
    { code: 'expect.assertions(1)' },
    { code: 'expect(true).toBe(...true)' },
    { code: 'expect()' },
    { code: 'expect({}).toStrictEqual({})' },
    { code: 'expect(a === b).toBe(true)' },
    { code: 'expect(a !== 2).toStrictEqual(true)' },
    { code: 'expect(a === b).not.toEqual(true)' },
    { code: 'expect(a !== "string").toStrictEqual(true)' },
    { code: 'expect(5 != a).toBe(true)' },
    { code: 'expect(a == "string").toBe(true)' },
    { code: 'expect(a == "string").not.toBe(true)' },
    { code: 'expect(value > 1)[matcher].toBe(true);' },
    { code: 'expect(value > 1)[foo()].toBe(true);' },
  ],
  invalid: [
    {
      code: 'expect(value > 1).toBe(true).toString();',
      output: 'expect(value).toBeGreaterThan(1).toString();',
      errors: [
        {
          messageId: 'useToBeComparison',
          data: { preferredMatcher: 'toBeGreaterThan' },
          column: 19,
          line: 1,
        },
      ],
    },
    {
      code: 'expect(value > 1).toBe(true).foo(false);',
      output: 'expect(value).toBeGreaterThan(1).foo(false);',
      errors: [
        {
          messageId: 'useToBeComparison',
          data: { preferredMatcher: 'toBeGreaterThan' },
          column: 19,
          line: 1,
        },
      ],
    },
  ],
});
