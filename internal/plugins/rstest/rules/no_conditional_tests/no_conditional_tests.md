# no-conditional-tests

## Rule Details

Disallows registering a test or a suite from inside an `if` statement. A test
that only exists on some runs is a test that silently stops covering anything
when the condition flips, and the report shows a shrinking suite rather than a
failure.

A registration is reported when it is reached through the `then` branch or
the `else` branch of an `if`; a call in the `if`'s own condition runs every
time and is not reported. Only `if` is reported — conditional expressions,
`switch`, and the logical operators are left alone. Hooks such as
`beforeEach` are not registrations and are not reported.

The rule looks only as far as the nearest enclosing function, so a
registration wrapped in a helper is attributed to wherever that helper is
called, not to an unrelated `if` the helper happens to sit under; the same
boundary means a nested pair such as
`if (x) { describe('a', () => { test('b', fn) }) }` reports only the outer
`describe`. Conditions written inside a test body are covered separately by
`rstest/no-conditional-in-test`.

A conditionally-run test should be registered with `test.skipIf(condition)`
or `test.runIf(condition)` instead, so the suite keeps its shape and the
runner decides at execution time whether to run it. Both modifiers are also
available on `describe`.

## Incorrect

```ts
if (process.env.CI) {
  test('uploads the report', async () => {
    await expect(upload()).resolves.toBe(true);
  });
}
```

## Correct

```ts
test.skipIf(!process.env.CI)('uploads the report', async () => {
  await expect(upload()).resolves.toBe(true);
});
```
