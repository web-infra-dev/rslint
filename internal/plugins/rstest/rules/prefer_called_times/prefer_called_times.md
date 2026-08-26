# prefer-called-times

## Rule Details

Requires call-count assertions to be spelled `expect(fn).toHaveBeenCalledTimes(1)` rather than `expect(fn).toHaveBeenCalledOnce()`. Both assert the same thing, and writing the count out keeps every call-count assertion in a file in one form, so changing an expected count becomes an edit to a number instead of a switch to a different matcher.

The matcher has to be called. `expect(fn).toHaveBeenCalledOnce` asserts nothing, so it is left alone, and so is `expect.toHaveBeenCalledOnce()`, which never received a value to count calls on. A matcher named by a computed key, `expect(fn)[matcherName]()`, is not reported either, because which assertion runs is only known at runtime. Chai's `calledOnce` is a property rather than a matcher call, and the assertion that takes a count in that style is `callCount(1)`, so `expect(fn).to.have.been.calledOnce` is outside this rule.

Every `expect` source is recognized: globals, named and renamed imports, `require` destructuring, namespace imports, whole-module `require`, `import.meta.rstest`, Playwright integrations, and the `expect` a test callback receives through its test context. An `expect` from another assertion library, or a local variable that shadows `expect`, is not reported.

## Incorrect

```ts
test('notifies the listener once', () => {
  emitter.emit('ready');

  expect(listener).toHaveBeenCalledOnce();
});
```

## Correct

```ts
test('notifies the listener once', () => {
  emitter.emit('ready');

  expect(listener).toHaveBeenCalledTimes(1);
});
```

## Autofix

Renames the matcher to `toHaveBeenCalledTimes` and inserts the count `1` into its argument list. Nothing else in the assertion changes: the expect root as written at the call site, `expect.soft`, the second `message` argument of `expect(actual, message)`, the modifier chain, the accessor's own quoting, and the surrounding line breaks and comments are all left as they are.
