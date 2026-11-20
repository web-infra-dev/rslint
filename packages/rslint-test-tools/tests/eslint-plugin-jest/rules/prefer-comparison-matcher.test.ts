// Upstream: eslint-plugin-jest@29.16.1, plus existing Rslint regressions.
import { RuleTester, type ValidTestCase } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

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
        message: `Prefer using \`${preferredMatcher}\` instead`,
        column: 18 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 18 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1,).${equalityMatcher}(true,);`,
    output: `expect(value,).${preferredMatcher}(1,);`,
    languageOptions: { ecmaVersion: 2017 },
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcher}\` instead`,
        column: 19 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 19 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)['${equalityMatcher}'](true);`,
    output: `expect(value).${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcher}\` instead`,
        column: 18 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 18 + operator.length + equalityMatcher.length + 2,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).resolves.${equalityMatcher}(true);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcher}\` instead`,
        column: 27 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 27 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).${equalityMatcher}(false);`,
    output: `expect(value).${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcherWhenNegated}\` instead`,
        column: 18 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 18 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)['${equalityMatcher}'](false);`,
    output: `expect(value).${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcherWhenNegated}\` instead`,
        column: 18 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 18 + operator.length + equalityMatcher.length + 2,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).resolves.${equalityMatcher}(false);`,
    output: `expect(value).resolves.${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcherWhenNegated}\` instead`,
        column: 27 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 27 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).not.${equalityMatcher}(true);`,
    output: `expect(value).${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcherWhenNegated}\` instead`,
        column: 22 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 22 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)['not'].${equalityMatcher}(true);`,
    output: `expect(value).${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcherWhenNegated}\` instead`,
        column: 25 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 25 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).resolves.not.${equalityMatcher}(true);`,
    output: `expect(value).resolves.${preferredMatcherWhenNegated}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcherWhenNegated}\` instead`,
        column: 31 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 31 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).not.${equalityMatcher}(false);`,
    output: `expect(value).${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcher}\` instead`,
        column: 22 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 22 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1).resolves.not.${equalityMatcher}(false);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcher}\` instead`,
        column: 31 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 31 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)["resolves"].not.${equalityMatcher}(false);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcher}\` instead`,
        column: 34 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 34 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)["resolves"]["not"].${equalityMatcher}(false);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcher}\` instead`,
        column: 37 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 37 + operator.length + equalityMatcher.length,
      },
    ],
  },
  {
    code: `expect(value ${operator} 1)["resolves"]["not"]['${equalityMatcher}'](false);`,
    output: `expect(value).resolves.${preferredMatcher}(1);`,
    errors: [
      {
        messageId: 'useToBeComparison',
        message: `Prefer using \`${preferredMatcher}\` instead`,
        column: 37 + operator.length,
        line: 1,
        endLine: 1,
        endColumn: 37 + operator.length + equalityMatcher.length + 2,
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

const operators = [
  ['>', 'toBeGreaterThan', 'toBeLessThanOrEqual'],
  ['<', 'toBeLessThan', 'toBeGreaterThanOrEqual'],
  ['>=', 'toBeGreaterThanOrEqual', 'toBeLessThan'],
  ['<=', 'toBeLessThanOrEqual', 'toBeGreaterThan'],
] as const;
const equalityMatchers = ['toBe', 'toEqual', 'toStrictEqual'];
const extraValid: ValidTestCase[] = [
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
  { code: 'expect().toStrictEqual({})' },
];

const comparisonCases = operators.map(
  ([operator, preferred, negated], run) => ({
    valid: [
      { code: 'expect()' },
      { code: 'expect({}).toStrictEqual({})' },
      { code: `expect(value).${preferred}(1);` },
      { code: `expect(value).${negated}(1);` },
      { code: `expect(value).not.${preferred}(1);` },
      { code: `expect(value).not.${negated}(1);` },
      ...equalityMatchers.flatMap((matcher) =>
        generateValidStringLiteralCases(operator, matcher).map((code) => ({
          code,
        })),
      ),
    ].map((item, index) => ({ name: `run ${run} valid ${index}`, ...item })),
    invalid: equalityMatchers
      .flatMap((matcher) =>
        generateInvalidCases(operator, matcher, preferred, negated),
      )
      .map((item, index) => ({ name: `run ${run} invalid ${index}`, ...item })),
  }),
);

// Run the complete matrix once with upstream's ES2015 default. The trailing
// argument-comma cases retain their explicit ES2017 override.
const config = fixtureConfig();
new RuleTester({
  ...config,
  languageOptions: { ...config.languageOptions, ecmaVersion: 2015 },
}).run('prefer-comparison-matcher', {
  valid: [
    ...comparisonCases.flatMap((cases) => cases.valid),
    ...extraValid.map((item, index) => ({
      name: `extra valid ${index}`,
      ...item,
    })),
  ],
  invalid: comparisonCases.flatMap((cases) => cases.invalid),
});

// Default configuration needs a smoke check, not a second complete matrix.
new RuleTester(fixtureConfig()).run('prefer-comparison-matcher', {
  valid: [
    {
      name: 'default configuration valid control',
      code: 'expect(value).toBeGreaterThan(1);',
    },
  ],
  invalid: [
    {
      ...generateInvalidCases(
        '>',
        'toBe',
        'toBeGreaterThan',
        'toBeLessThanOrEqual',
      )[0],
      name: 'default configuration autofix control',
    },
  ],
});
