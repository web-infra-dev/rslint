# valid-expect-in-promise

Require promise chains containing Rstest assertions to be returned, awaited, or consumed by a safe promise sink.

Rstest has two official assertion interfaces:

```ts
expect(actual).toBe(expected);
assert.equal(actual, expected);
```

`expect` is Rstest's extended `@vitest/expect` and Chai interface. `assert` is Chai's function-style interface exported by `@rstest/core`. They share the Chai assertion engine but are not aliases; `assert` does not contribute to `expect.assertions()` bookkeeping. Both throw assertion errors, so both have the same floating-promise risk:

```ts
test('loads a value', () => {
  load().then(value => {
    assert.equal(value, 'ready');
  });
});
```

Return or await the chain:

```ts
test('loads a value', async ({ expect }) => {
  await load().then(value => {
    expect(value).toBe('ready');
  });
});
```

The rule recognizes Rstest globals, `@rstest/core` imports and requires, namespace objects, `import.meta.rstest`, TestContext `expect`, and `@rstest/playwright` `expect`. It does not treat Node, direct Chai, or locally defined `assert` functions as Rstest assertions.

Rstest does not support Jest-style `done` callbacks. A regular test callback's first argument is `TestContext`; `test.for` receives context as its second callback argument, while `test.each` receives case values only. All of these callbacks are analyzed.

The rule follows promises stored in local bindings, including statically mappable array and object destructuring, and requires every reachable path to consume them. `Promise.resolve` and `Promise.all` preserve assertion failure. `Promise.reject` and `Promise.allSettled` are not safe sinks.

An awaited chain is safe only when its rejection can escape the test callback. An enclosing `catch` that may swallow the rejection, or a `finally` block that may `return`, `break`, or `continue`, does not count as safe consumption. A `catch` that always rethrows preserves the failure.

This rule has no options and does not provide an autofix.
