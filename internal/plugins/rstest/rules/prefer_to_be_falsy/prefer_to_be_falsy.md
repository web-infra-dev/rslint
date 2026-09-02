# prefer-to-be-falsy

## Rule Details

Requires an assertion that a value is `false` to be spelled `expect(value).toBeFalsy()` rather than compared against the literal with `toBe`, `toEqual` or `toStrictEqual`. Most values checked this way are flags a function returns, and the named matcher says what the test is about — the value is falsy — instead of asserting an exact identity the test rarely depends on.

The literal has to be written out as the matcher's only argument. `expect(value).toBe(0)`, `expect(value).toBe(!1)` and `expect(value).toBe(flag)` are all left alone, and so is a matcher that takes no argument at all, such as `expect(value).toBe()`. A type assertion around the literal is transparent, so `expect(value).toBe(false as boolean)` is reported. The matcher also has to be called: `expect(value).toBe` asserts nothing, so it is left alone, and so is `expect.toBe(false)`, which never received a value to assert on. A matcher named by a computed key, `expect(value)[matcherName](false)`, is not reported either, because which assertion runs is only known at runtime. Chai's falsy assertions are properties rather than matcher calls, and its equality form is named `equal`, so `expect(value).to.be.false` and `expect(value).to.equal(false)` are both outside this rule.

Modifiers are kept as written, so `expect(value).not.toBe(false)` becomes `expect(value).not.toBeFalsy()`. That asserts something wider than before — every truthy value passes, not just "anything other than `false`" — which is the assertion this rule is asking for.

Every `expect` source is recognized: globals, named and renamed imports, `require` destructuring, namespace imports, whole-module `require`, `import.meta.rstest`, Playwright integrations, and the `expect` a test callback receives through its [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). An `expect` from another assertion library, or a local variable that shadows `expect`, is not reported.

## Incorrect

```ts
test('reports an empty configuration as invalid', () => {
  expect(isValidConfig({})).toBe(false);
});
```

## Correct

```ts
test('reports an empty configuration as invalid', () => {
  expect(isValidConfig({})).toBeFalsy();
});
```

## Autofix

Renames the matcher to `toBeFalsy` and empties its argument list. Nothing else in the assertion changes: the expect root as written at the call site, `expect.soft`, the second `message` argument of `expect(actual, message)`, the modifier chain, the accessor's own quoting, and the surrounding line breaks and comments are all left as they are. The assertion is reported without a fix when a comment sits between the parentheses, as in `expect(value).toBe(/* the flag */ false)`, since emptying the argument list would delete it.
