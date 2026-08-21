# max-expects

## Rule Details

Enforce a maximum number of assertion calls in each Rstest test body. This
rule counts real assertion factories such as `expect(value)`,
`expect.soft(value)`, `expect.poll(fn)`, and `expect.element(locator)`. Once a
test exceeds the configured maximum, each additional assertion is reported.

Examples of **incorrect** code for this rule:

```ts
test("case", () => {
  expect(a).toBe(1);
  expect(b).toBe(2);
  expect(c).toBe(3);
  expect(d).toBe(4);
  expect(e).toBe(5);
  expect(f).toBe(6);
});
```

Examples of **correct** code for this rule:

```ts
test("case", () => {
  expect(a).toBe(1);
  expect(b).toBe(2);
  expect(c).toBe(3);
});

test.for(rows)("case", (row, context) => {
  context.expect(row.ok).toBe(true);
  context.expect(row.count).toBeGreaterThan(0);
});
```

Examples of **correct** code for this rule with `{ "max": 1 }`:

```json
{ "rstest/max-expects": ["error", { "max": 1 }] }
```

```ts
test("case", () => {
  expect(value).toBe(1);
});
```

## Rstest specifics

- `expect.assertions(2)`, `expect.hasAssertions()`, `expect.any(String)`, and
  other static `expect` helpers do not count as assertions.
- Broken or incomplete chains such as `expect(value)` and `expect(value).toBe`
  do not count; those are covered by `rstest/valid-expect`.
- Chai property and chained assertions still count once per assertion factory:
  `expect(value).to.be.ok` and
  `expect(value).to.be.a("string").and.not.be.empty` each count as one.
- Assertions outside test callbacks are ignored. Detached helpers inside a test
  get their own count and do not leak into sibling tests.

## Options

```json
{
  "rstest/max-expects": ["error", { "max": 5 }]
}
```

- `max` configures the maximum number of assertions allowed per test body.
  The default is `5`.

This rule does not provide an autofix.

## Original Documentation

- [eslint-plugin-jest: max-expects](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/max-expects.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/max-expects.ts)
