# require-awaited-expect-poll

## Rule Details

Requires the promise returned by `expect.poll(...)` or `expect.element(...)` to be handled. `expect.poll(fn)` re-evaluates `fn` until its matcher passes or the timeout expires, and every matcher `expect.element(locator)` exposes is `async`, so dropping the promise lets the test finish before the assertion does — a failing assertion then either surfaces against an unrelated test or does not surface at all.

An assertion counts as handled when it is awaited or returned, or when its value is passed on somewhere that can settle it later: a concise arrow body, a variable initializer, the right side of an assignment, a call argument — which covers `Promise.all([...])` and `Promise.allSettled([...])` — an array element, a `yield` operand, or an object property value. A chain whose matcher never runs, such as `expect.poll(fn).toBeVisible` or `expect.poll(fn)`, is not reported.

Every `expect` source is recognized: globals, named and renamed imports, `require` destructuring, namespace imports, whole-module `require`, `import.meta.rstest`, Playwright integrations, and the `expect` a test callback receives through its [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). An `expect` from another assertion library, or a local variable that shadows `expect`, is not reported.

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
