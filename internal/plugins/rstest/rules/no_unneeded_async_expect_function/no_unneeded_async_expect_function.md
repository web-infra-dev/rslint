# no-unneeded-async-expect-function

## Rule Details

Assert on the call itself instead of on an async function that only awaits it. `rejects` already calls a function handed to `expect()` and awaits what it returns, so an `async` wrapper whose whole body is `await someCall()` adds a layer without changing what is asserted. Written directly, the assertion also reads as what it checks: the promise `someCall()` returns.

The rule only looks at assertions that go through `resolves` or `rejects`. Every other chain asserts on the function object — `expect(handler).toBeInstanceOf(Function)` and `expect(handler).toThrow()` both describe the function, not a promise it would return — so the wrapper is meaningful there and is left alone. It reports `expect()` and `expect.soft()`, which assert on the value passed to them, and not `expect.poll()`, which retries a callback, or `expect.element()`, which asserts on a browser locator. Global `expect`, named and renamed imports, namespace imports and CommonJS bindings from `@rstest/core`, `rstack/test` and `@rstest/playwright`, `import.meta.rstest.expect`, and the `expect` from a test context are all recognized. The rule needs no type information.

The wrapper has to be an `async` function expression or arrow function whose body is a single awaited call, written either as a block with one statement or as a concise arrow body. Anything else keeps the wrapper doing work: a second statement, a declaration, a loop, an `await` of something that is not a call, an array or object built around the awaited value, and an async generator, whose call returns an async iterator rather than a promise.

## Incorrect

```ts
import { expect, test } from '@rstest/core';

test('rejects an unknown user', async () => {
  await expect(async () => {
    await loadProfile('unknown');
  }).rejects.toThrow('user not found');
});
```

## Correct

```ts
import { expect, test } from '@rstest/core';

test('rejects an unknown user', async () => {
  await expect(loadProfile('unknown')).rejects.toThrow('user not found');
});
```

## Suggestions

The rule reports without an autofix and offers the rewrite as a suggestion, because the two spellings differ when the awaited call throws synchronously instead of returning a rejected promise. Inside the wrapper such a throw becomes a rejection and the assertion passes; once the wrapper is gone the call runs while `expect()`'s arguments are evaluated, and the error escapes the assertion and fails the test. Whether a call can throw synchronously is not decidable from the call itself, so the change is left for you to accept.

The suggestion replaces the whole wrapper with the awaited call, keeping the call exactly as written, including optional calls and comments inside it, along with the parentheses and the second `expect()` argument around it.

It is withheld, and the assertion is reported on its own, when the unwrapped call would no longer mean the same thing:

- The wrapper declares parameters or type parameters, is a named function expression, or reads `this` or `arguments` while being a function expression. `rejects` calls the wrapper with no arguments and no receiver, so those names are bound by the wrapper and would resolve elsewhere — or nowhere — in the assertion's own scope. An arrow function takes `this` and `arguments` from the enclosing scope already, so it keeps the suggestion.
- `expect()` carries an explicit type argument, which describes the wrapper rather than the awaited value.
- A comment sits inside the wrapper but outside the awaited call, where the rewrite would delete it.
