# jest/valid-expect-in-promise

Require promise chains containing Jest `expect` calls to be returned, awaited, or consumed by a safe promise sink.

```js
test('loads a value', () => {
  load().then(value => {
    expect(value).toBe('ready');
  });
});
```

The test can finish before the handler runs. Return or await the chain:

```js
test('loads a value', async () => {
  await load().then(value => {
    expect(value).toBe('ready');
  });
});
```

The rule follows promises stored in local bindings, including statically mappable array and object destructuring, and requires every reachable path to consume them. `Promise.resolve`, `Promise.all`, and single-input `Promise.any` / `Promise.race` preserve assertion failure. `Promise.reject` does not adopt its argument, and `Promise.allSettled` converts assertion failure into a fulfilled result, so neither is a safe sink.

Jest callbacks using `done` are not analyzed because the callback can coordinate promise completion explicitly. Named callbacks and promise-handler callbacks are analyzed when their relationship to a test is statically known.

This implementation intentionally fixes upstream cases where unrelated assignments stop promise tracking, named callbacks are missed, nested handlers are missed, or `Promise.reject` / `Promise.allSettled` are incorrectly accepted.

This rule has no options and does not provide an autofix.
