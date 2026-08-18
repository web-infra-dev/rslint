# valid-expect-in-promise

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

The rule follows promises stored in local bindings, including statically mappable array and object destructuring, and requires every reachable path to consume them. `Promise.resolve` and `Promise.all` preserve assertion failure. `Promise.reject` does not adopt its argument, and `Promise.allSettled` converts assertion failure into a fulfilled result, so neither is a safe sink.

A logical assignment stores the chain like a plain assignment does, and its right-hand side is only evaluated when the store happens, so `pending ||= load().then(...)` followed by `await pending` is consumed. In the other direction a promise is truthy and non-nullish, so a later `||=` or `??=` cannot replace the promise a binding already holds, while `&&=` always does and discards it.

The chain also reaches its binding through `||`, `??`, `&&`, either branch of a conditional, and the right-hand side of a comma, because each of those evaluates the chain on exactly the paths where the binding receives it. `let pending = fallback || load().then(...)` is therefore tracked, and so is the longhand `pending = pending || load().then(...)`, which keeps the promise the binding already holds for the same reason `||=` does.

An array literal in a binding holds the chain, and only an element-wise consumption of that binding reaches it: `await Promise.all(pending)` and `for await (const settled of pending)` consume it, while `await pending` awaits the array itself and leaves the chain floating.

A chain is consumed only when it is awaited, returned, wrapped in `Promise.resolve` or `Promise.all`, consumed element-wise by `for await`, or handed to an `expect(...).resolves` / `.rejects` sink. Every other position leaves it floating, including several the upstream rule does not report: `Promise.race` and `Promise.any`, any other wrapper call such as `helper(chain)`, `void chain` and `throw chain`, a property of an object literal that is not destructured straight into a binding, an array literal that is never consumed element-wise, a tagged template interpolation, and a compound arithmetic assignment such as `count += chain`, which stores no promise at all.

An awaited chain is safe only when its rejection can escape the test callback. An enclosing `catch` that may swallow the rejection, or a `finally` block that may `return`, `break`, or `continue`, does not count as safe consumption. A `catch` that always rethrows preserves the failure.

Jest callbacks using `done` are not analyzed because the callback can coordinate promise completion explicitly. Named callbacks and promise-handler callbacks are analyzed when their relationship to a test is statically known.

This implementation intentionally fixes upstream cases where unrelated assignments stop promise tracking, named callbacks are missed, nested handlers are missed, or `Promise.reject` / `Promise.allSettled` are incorrectly accepted.

This rule has no options and does not provide an autofix.

## Original Documentation

- [eslint-plugin-jest: valid-expect-in-promise](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/valid-expect-in-promise.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/valid-expect-in-promise.ts)
