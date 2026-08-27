# no-conditional-tests

## Rule Details

Disallows registering a test or a suite from inside an `if` statement. A test
that only exists on some runs is a test that silently stops covering anything
when the condition flips, and the report shows a shrinking suite rather than a
failure.

A registration is reported when it is reached through the `then` branch or the
`else` branch of an `if`. Calls in the `if`'s own condition are not reported —
they run every time. Only `if` is reported; conditional expressions, `switch`,
and the logical operators are left alone.

The rule stops looking at the first enclosing function, so a registration
wrapped in a helper function is attributed to wherever that helper is called,
not to an `if` the helper happens to sit under. The same boundary means a
nested pair reports only the outermost registration: in
`if (x) { describe('a', () => { test('b', fn) }) }` only the `describe` is
reported, because the inner `test` belongs to the suite callback.

Hooks (`beforeEach`, `afterAll`, and the rest) are not reported.

Conditions written *inside* a test body are a different concern, covered by
`rstest/no-conditional-in-test`.

## Incorrect

```ts
if (process.env.CI) {
  test('uploads the report', async () => {
    await expect(upload()).resolves.toBe(true);
  });
}
```

## Correct

Rstest registers the test either way and decides at run time whether to execute
it, so the suite keeps its shape:

```ts
test.skipIf(!process.env.CI)('uploads the report', async () => {
  await expect(upload()).resolves.toBe(true);
});
```

`test.runIf(condition)` expresses the same thing from the other side, and both
modifiers are available on `describe` as well.
