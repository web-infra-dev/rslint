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
- A test body and a lifecycle hook body each carry their own count:
  `beforeEach`, `beforeAll`, `afterEach`, and `afterAll` are test code and are
  limited the same way a test is.
- A callback the rule cannot resolve is still counted when it is written inside
  the registration's callback argument, wherever in that argument it sits:
  `test("a", fakeAsync(() => …))`, `test("a", new Wrapper(() => …))`,
  `test("a", wrap({ cb: () => … }))`, and `test("a", cond ? () => … : null)`
  all count. A function held in an options object — the `{ retry: 2 }` argument
  of the `(name, options, callback)` overload — is not a callback and is not
  counted.
- A callback written outside the registration and reached through a member
  expression is not counted: in `const suite = { run: () => … };
  test("a", suite.run)`, and likewise for a class property, nothing at the
  registration ties the function to it.
- Assertions at the top level of a file, or directly in a `describe` body, are
  ignored: they do not belong to a test body.
- A detached helper inside a test gets its own count and does not leak into
  sibling tests.

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
