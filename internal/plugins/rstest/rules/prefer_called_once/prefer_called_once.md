# prefer-called-once

## Rule Details

Requires an assertion that a mock was called exactly once to be spelled `expect(fn).toHaveBeenCalledOnce()` rather than `expect(fn).toHaveBeenCalledTimes(1)` or its deprecated alias `expect(fn).toBeCalledTimes(1)`. All three assert the same thing, and the named matcher puts "once" in the matcher instead of in a count the reader has to check.

Only a literal `1` counts, so `toHaveBeenCalledTimes(0)`, `toHaveBeenCalledTimes(2)` and `toHaveBeenCalledTimes(count)` are all left alone, and so is a count written as the BigInt `1n` or as an expression such as `+1`. The matcher also has to be called: `expect(fn).toHaveBeenCalledTimes` asserts nothing, so it is left alone, and so is `expect.toHaveBeenCalledTimes(1)`, which never received a value to count calls on. A matcher named by a computed key, `expect(fn)[matcherName](1)`, is not reported either, because which assertion runs is only known at runtime. Chai's `calledOnce` is a property rather than a matcher call, and the count-taking form of that assertion is `callCount(1)`, so `expect(fn).to.have.been.calledOnce` and `expect(fn).callCount(1)` are both outside this rule.

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

Renames the matcher to `toHaveBeenCalledOnce` and empties its argument list. Nothing else in the assertion changes: the expect root as written at the call site, `expect.soft`, the second `message` argument of `expect(actual, message)`, the modifier chain, the accessor's own quoting, and the surrounding line breaks and comments are all left as they are. The assertion is reported without a fix when a comment sits between the parentheses, as in `expect(fn).toHaveBeenCalledTimes(/* exactly one */ 1)`, since emptying the argument list would delete it.
