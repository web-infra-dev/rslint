# prefer-called-once

## Rule Details

Requires an assertion that a mock was called exactly once to be spelled `expect(fn).toHaveBeenCalledOnce()` rather than `expect(fn).toHaveBeenCalledTimes(1)` or its deprecated alias `expect(fn).toBeCalledTimes(1)`. All three assert the same thing, and the named matcher puts "once" in the matcher instead of in an argument the reader has to look at.

Only a literal `1` counts. `toHaveBeenCalledTimes(0)`, `toHaveBeenCalledTimes(2)` and `toHaveBeenCalledTimes(count)` are all left alone, and so is a count written as the BigInt `1n` or as an expression such as `+1`.

The matcher has to be called. `expect(fn).toHaveBeenCalledTimes` asserts nothing, so it is left alone, and so is `expect.toHaveBeenCalledTimes(1)`, which never received a value to count calls on. A matcher named by a computed key, `expect(fn)[matcherName](1)`, is not reported either, because which assertion runs is only known at runtime.

Every `expect` source is recognized: globals, named and renamed imports, `require` destructuring, namespace imports, whole-module `require`, `import.meta.rstest`, Playwright integrations, and the `expect` a test callback receives through its [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). An `expect` from another assertion library, or a local variable that shadows `expect`, is not reported.

## Incorrect

```ts
test('notifies the listener once', () => {
  emitter.emit('ready');

  expect(listener).toHaveBeenCalledTimes(1);
});
```

## Correct

```ts
test('notifies the listener once', () => {
  emitter.emit('ready');

  expect(listener).toHaveBeenCalledOnce();
});
```

## Autofix

Renames the matcher to `toHaveBeenCalledOnce` and empties its argument list. Nothing else in the assertion changes: the expect root as written at the call site, `expect.soft`, the second `message` argument of `expect(actual, message)`, the modifier chain, the accessor's own quoting, and the surrounding line breaks and comments are all left as they are.

Emptying the argument list would also delete a comment written between the parentheses, so an assertion such as `expect(fn).toHaveBeenCalledTimes(/* exactly one */ 1)` is reported without a fix and left for you to rewrite.

## Known gaps

Chai's call-count assertion, `expect(fn).callCount(1)`, is not reported. Its "once" spelling is the `calledOnce` property rather than a matcher call, so rewriting it is a different edit than the one this rule makes.

## Relationship to other rules

This rule is the exact inverse of `rstest/prefer-called-times`, which rewrites `toHaveBeenCalledOnce()` back into `toHaveBeenCalledTimes(1)`. Enable one or the other, never both: with both on, every call-count assertion is reported no matter how it is written, and the two autofixes undo each other.

Enabling this rule together with `rstest/no-alias-methods` is fine. Both rewrite towards the canonical matcher names, so they agree on the result.

## Differences from ESLint

- `expect(fn).toBeCalledTimes(1)` is rewritten to `expect(fn).toHaveBeenCalledOnce()`. ESLint writes `expect(fn).toBeCalledOnce()`, a matcher Rstest's assertion library does not define, so the fixed code throws where the original passed.
- The message always names `toHaveBeenCalledOnce()`, whichever of the two count matchers was written. ESLint's message names whichever matcher its fix would write.
- `expect(fn).toHaveBeenCalledTimes(1,)` is rewritten to `expect(fn).toHaveBeenCalledOnce()`. ESLint removes only the `1` and leaves `expect(fn).toHaveBeenCalledOnce(,)`, which does not parse.
- `expect(fn)['toBeCalledTimes'](1)` keeps its quoting and becomes `expect(fn)['toHaveBeenCalledOnce']()`. ESLint drops the quotes and produces `expect(fn)[toHaveBeenCalledOnce]()`, which reads a variable instead of naming the matcher.
- `expect(fn).toHaveBeenCalledTimes /* why */ (1)` is fixed without disturbing the comment.
- `(expect(fn)?.toHaveBeenCalledTimes)(1)` is reported. ESLint does not report an assertion whose optional chain is parenthesized this way.
