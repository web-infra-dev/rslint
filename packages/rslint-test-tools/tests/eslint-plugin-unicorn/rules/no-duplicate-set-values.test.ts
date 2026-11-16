// Ported from eslint-plugin-unicorn v80.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/test/no-duplicate-set-values.js
import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

const duplicate = (value: string) => ({
  messageId: 'no-duplicate-set-values',
  message: `Remove duplicate value \`${value}\` from the Set.`,
});

const invalid = (code: string, ...values: string[]): InvalidTestCase => ({
  code,
  errors: values.map((value) => duplicate(value)),
});

const invalidTS = (code: string, ...values: string[]): InvalidTestCase => ({
  ...invalid(code, ...values),
  filename: 'src/virtual.ts',
});

const validCases: ValidTestCase[] = [
  'new Set([1, 2, "1"])',
  'new Set([{}, {}])',
  'new Set([[], []])',
  'new Set([/foo/, /foo/])',
  'new Set([Symbol(), Symbol()])',
  'new Set(foo)',
  'Set([1, 1])',
  'new NotSet([1, 1])',
  'new globalThis.Set([1, 1])',
  'new Set([foo(), foo()])',
  'new Set([foo.bar, foo.baz])',
  `const object = {value: 1};
Object.defineProperty(object, 'value', {get() { return 2; }});
new Set([object.value, 1]);`,
  'new Set([foo, ...bar, baz])',
  `const foo = {};
new Set([foo, {}]);`,
  `const foo = 1;
const bar = 2;
new Set([foo, bar]);`,
].map((code) => ({ code }));

const invalidCases: InvalidTestCase[] = [
  invalid('new Set([1, 2, 1])', '1'),
  invalid('new Set(["foo", "bar", "foo"])', '"foo"'),
  invalid('new Set([null, null])', 'null'),
  invalid('new Set([undefined, undefined])', 'undefined'),
  invalid('new Set([, undefined])', 'undefined'),
  invalid('new Set([undefined, ,])', 'undefined'),
  invalid('new Set([,,])', 'undefined'),
  invalid('new Set([-1, -1])', '-1'),
  invalid('new Set([NaN, NaN])', 'NaN'),
  invalid('new Set([0, -0])', '-0'),
  invalid('new Set([1, 1, 2, 2])', '1', '2'),
  invalid('new Set([1 + 1, 2])', '2'),
  invalid('new Set([`foo`, "foo"])', '"foo"'),
  invalid('new Set([foo, foo])', 'foo'),
  invalid('new Set([foo.bar, foo.bar])', 'foo.bar'),
  invalid('new Set([foo["bar"], foo.bar])', 'foo.bar'),
  invalid('new Set([this.foo, this.foo])', 'this.foo'),
  invalid('new Set([foo, ...bar, foo])', 'foo'),
  invalid(
    `const foo = 2;
new Set([foo, 2]);`,
    '2',
  ),
  invalid(
    `const foo = 'bar';
new Set([foo, 'bar']);`,
    "'bar'",
  ),
  invalid(
    `const foo = undefined;
new Set([foo, undefined]);`,
    'undefined',
  ),
  invalid(
    `const foo = {};
new Set([foo, foo]);`,
    'foo',
  ),

  // TypeScript wrappers around the array literal
  invalidTS('new Set([1, 2, 1] as const);', '1'),
  invalidTS('new Set([1, 1]!);', '1'),
  invalidTS('new Set([1, 2, 1] satisfies number[]);', '1'),
];

new RuleTester().run('no-duplicate-set-values', {} as never, {
  valid: validCases,
  invalid: invalidCases,
});
