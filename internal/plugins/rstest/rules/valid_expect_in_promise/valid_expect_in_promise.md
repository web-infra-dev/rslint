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

A logical assignment stores the chain like a plain assignment does, and its right-hand side is only evaluated when the store happens, so `pending ||= load().then(...)` followed by `await pending` is consumed. In the other direction a promise is truthy and non-nullish, so a later `||=` or `??=` cannot replace the promise a binding already holds, while `&&=` always does and discards it.

The chain also reaches its binding through `||`, `??`, `&&`, either branch of a conditional, and the right-hand side of a comma, because each of those evaluates the chain on exactly the paths where the binding receives it. `let pending = fallback || load().then(...)` is therefore tracked, and so is the longhand `pending = pending || load().then(...)`, which keeps the promise the binding already holds for the same reason `||=` does.

An array literal in a binding holds the chain, and only an element-wise consumption of that binding reaches it: `await Promise.all(pending)` and `for await (const settled of pending)` consume it, while `await pending` awaits the array itself and leaves the chain floating.

A chain is consumed only when it is awaited, returned, wrapped in `Promise.resolve` or `Promise.all`, consumed element-wise by `for await`, or handed to an `expect(...).resolves` / `.rejects` sink. Every other position leaves it floating: `Promise.race` and `Promise.any`, any other wrapper call such as `helper(chain)`, `void chain` and `throw chain`, a property of an object literal that is not destructured straight into a binding, an array literal that is never consumed element-wise, a tagged template interpolation, and a compound arithmetic assignment such as `count += chain`, which stores no promise at all.

An awaited chain is safe only when its rejection can escape the test callback. An enclosing `catch` that may swallow the rejection, or a `finally` block that may `return`, `break`, or `continue`, does not count as safe consumption. A `catch` that always rethrows preserves the failure.

This rule has no options and does not provide an autofix.
