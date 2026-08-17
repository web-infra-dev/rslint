# valid-expect

## Rule Details

Enforce valid `expect()` usage in Rstest. `expect` must be called with the
right number of arguments, must be followed by a matcher, and asynchronous
assertions must be awaited or returned so their failures are not lost.

Examples of incorrect code:

```ts
expect();                       // notEnoughArgs
expect(value);                  // matcherNotFound
expect(value).toBe;             // matcherNotCalled
expect(value).notAModifier();   // modifierUnknown

test("async", async () => {
  expect(promise).resolves.toBe(1); // asyncMustBeAwaited
});
```

Examples of correct code:

```ts
expect(value).toBe(1);
expect(value).not.toBe(2);
expect(value, "message").toBe(1);   // second message argument is allowed

test("async", async () => {
  await expect(promise).resolves.toBe(1);
  await expect(promise).rejects.toThrow();
});
```

## Rstest specifics

rstest's `expect` comes from `@vitest/expect` + chai, so this rule follows the
vitest behavior where it differs from jest:

- **A second argument to `expect` is allowed** when it is a message string or
  template literal (`expect(value, "msg")`), and `expect.poll(fn, options)` /
  `expect.element(el, options)` accept an options object. These do not trigger
  `tooManyArgs`.
- **Chai property matchers are valid** without a call: `expect(value).to.be.ok`,
  `expect(spy).to.have.been.called`. They are not reported as `matcherNotCalled`.
- **Forms with no assertion factory carry no assertion**: `expect.assertions(1)`,
  `expect.hasAssertions()`, asymmetric matchers such as `expect.any(Number)` and
  bare chains such as `expect.resolves.toBe(1)` or `expect.toResolve()` are not
  subject to argument or await checks.

`expect` is recognized from globals, `@rstest/core` imports and aliases,
`require`, namespace access, `import.meta.rstest`, test-context `expect`
(`test('x', ({ expect }) => ...)`), and `@rstest/playwright`.

## Options

```json
{
  "rstest/valid-expect": [
    "error",
    {
      "alwaysAwait": false,
      "asyncMatchers": ["toReject", "toResolve"],
      "minArgs": 1,
      "maxArgs": 1
    }
  ]
}
```

- **`alwaysAwait`** — require every async assertion to be awaited, disallowing
  `return`.
- **`asyncMatchers`** — matcher names treated as asynchronous (must be awaited).
- **`minArgs`** / **`maxArgs`** — the required argument count for `expect`.

The rule provides an automatic fix that inserts `await` (and `async` on the
enclosing function) for async assertions that are not awaited.
