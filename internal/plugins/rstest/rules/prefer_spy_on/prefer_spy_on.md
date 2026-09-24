# prefer-spy-on

## Rule Details

This rule requires [`rs.spyOn()`](https://rstest.rs/api/runtime-api/rstest/mock-functions#rsspyon) instead of assigning [`rs.fn()`](https://rstest.rs/api/runtime-api/rstest/mock-functions#rsfn) to an object property. An assignment replaces the property for good: [`rs.restoreAllMocks()`](https://rstest.rs/api/runtime-api/rstest/mock-functions#rsrestoreallmocks) and the [`restoreMocks`](https://rstest.rs/config/test/restore-mocks) option cannot bring the original back, so the mock leaks into every later test that uses the object. A spy records the original method and restores it.

The rule reports a plain `=` assignment to a dotted or bracketed property when the assigned value is an `rs.fn()` call, including one followed by further configuration such as `.mockReturnValue()` and one wrapped in parentheses or a TypeScript `as`, `satisfies` or `!`. `rs.fn()` is recognized through `rs` and `rstest`, renamed imports from `@rstest/core` or `rstack/test`, namespace imports, `require`, and `import.meta.rstest.rs`; a local variable named `rs` is a different object and is not reported. The rule does not require type information.

Compound assignments such as `obj.method ??= rs.fn()` are not reported, because they install the mock only conditionally. Assignments to private fields are not reported either, because `rs.spyOn()` cannot reach them.

## Incorrect

```ts
test('shows the cached profile', async () => {
  api.fetchProfile = rs.fn().mockResolvedValue({ name: 'Ada' });
  Date.now = rs.fn(() => 1_700_000_000_000);

  await expect(loadProfile()).resolves.toMatchObject({ name: 'Ada' });
});
```

## Correct

```ts
afterEach(() => {
  rs.restoreAllMocks();
});

test('shows the cached profile', async () => {
  rs.spyOn(api, 'fetchProfile').mockResolvedValue({ name: 'Ada' });
  rs.spyOn(Date, 'now').mockImplementation(() => 1_700_000_000_000);

  await expect(loadProfile()).resolves.toMatchObject({ name: 'Ada' });
});
```

## Autofix

The autofix rewrites `object.method = rs.fn(implementation)` to `rs.spyOn(object, 'method').mockImplementation(implementation)`, calling `spyOn` on the same `rs`, `rstest`, alias or namespace that `fn` was called on. Configuration chained after `rs.fn()` stays attached to the spy. When `rs.fn()` has no implementation, the fix passes `() => undefined`: a spy without an implementation calls the original method, while `rs.fn()` returned `undefined`. When the chain sets `.mockImplementation()` right after `rs.fn()`, the implementation passed to `rs.fn()` is dropped.

The rewritten code differs from the assignment in three cases, which the fix does not detect:

- `rs.spyOn()` throws when the property does not exist or is not a function, so an assignment that created a new property fails after the fix.
- `mockReset()`, `rs.resetAllMocks()` and the [`resetMocks`](https://rstest.rs/config/test/reset-mocks) option return a spy to the original method, where `rs.fn(implementation)` returned to `implementation`.
- The spy is typed from the original method. In TypeScript, `() => undefined` does not type-check against a method whose return type excludes `undefined`, such as `Date.now`; give the spy an implementation that returns a valid value.

No autofix is offered when the rewrite would delete a comment, when `rs.fn()` is given more than one argument, or when the property is read from `super`.
