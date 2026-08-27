# require-awaited-expect-poll

## Rule Details

Requires the promise returned by an `expect.poll(...)` or an `expect.element(...)` assertion to be handled.

Both factories are asynchronous. `expect.poll(fn)` re-evaluates `fn` until its matcher passes or the timeout expires, and every matcher `expect.element(locator)` exposes is declared `async`. Dropping the promise they return means the test finishes before the assertion does, so a failing assertion either surfaces against an unrelated test or does not surface at all.

An assertion counts as handled when it is awaited or returned, and also when its value is passed on somewhere that can settle it later: a concise arrow body, a variable initializer, the right-hand side of an assignment, a call argument — which covers `Promise.all([...])` and `Promise.allSettled([...])` — an array element, a `yield` operand, or an object property value. Parentheses and TypeScript `as` / `satisfies` / `!` wrappers do not hide any of these. A comma expression evaluates to its last operand, so an assertion written last inside one is handled exactly when the comma expression itself is, while an assertion in any earlier position is reported.

Only assertions Rstest owns are reported. Every `expect` source is recognized: globals, named and renamed imports, `require` destructuring, namespace imports, whole-module `require`, `import.meta.rstest`, Playwright integrations, and the `expect` a test callback receives through its [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). An `expect` from another assertion library, or a local variable that shadows `expect`, is left alone, and so is a chain whose matcher never runs, such as `expect.poll(fn).toBeVisible` or `expect.poll(fn)`.

The reported range is the factory access itself — `expect.poll`, or `expect['poll']` including its closing bracket — and each chain is reported at most once.

## Incorrect

```ts
test('element exists', () => {
  asyncInjectElement();

  expect.poll(() => document.querySelector('.element')).toBeInTheDocument();
});
```

## Correct

```ts
test('element exists', async () => {
  asyncInjectElement();

  await expect
    .poll(() => document.querySelector('.element'))
    .toBeInTheDocument();
});
```

```ts
test('element exists', () => {
  asyncInjectElement();

  return expect
    .poll(() => document.querySelector('.element'))
    .toBeInTheDocument();
});
```

## Autofix

None. Whether a dropped assertion should be awaited, returned, or collected with the other promises around it is a decision about the test, not a mechanical rewrite.

## Differences

Compared with the Vitest rule of the same name:

- Only `await` and `return` count as handling the promise there. Here an assertion whose value is bound or passed on — `const p = expect.poll(fn).toBe(1)`, `await Promise.all([expect.poll(fn).toBe(1)])`, `() => expect.element(el).toBeVisible()` — is accepted instead of reported.
- `await (expect.poll(fn).toBe(1) as Promise<void>)` and `await (expect.poll(fn).toBe(1)!)` are accepted here; the Vitest rule reports them because the wrapper sits between the assertion and the `await`.
