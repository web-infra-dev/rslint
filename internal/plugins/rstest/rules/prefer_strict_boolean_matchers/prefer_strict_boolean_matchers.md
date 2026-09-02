# prefer-strict-boolean-matchers

## Rule Details

Requires an assertion about a boolean to be spelled `expect(value).toBe(true)` or `expect(value).toBe(false)` rather than `toBeTruthy()` or `toBeFalsy()`. The truthiness matchers coerce whatever they are given, so `''`, `0`, `null` and `undefined` all satisfy `toBeFalsy()`; asserting the literal keeps the test honest about the value it actually expects.

The matcher has to be called and has to be given nothing, since `toBeTruthy` and `toBeFalsy` take no arguments: `expect(value).toBeTruthy` asserts nothing and is left alone, and so is `expect.toBeTruthy()`, which never received a value to assert on. A matcher named by a computed key, `expect(value)[matcherName]()`, is not reported either, because which assertion runs is only known at runtime. Chai's truthiness assertions are properties rather than matcher calls, and their strict form is `equal(true)`, so `expect(value).to.be.ok` and `expect(value).to.be.true` are both outside this rule.

Modifiers are kept as written, so `expect(value).not.toBeTruthy()` becomes `expect(value).not.toBe(true)`. That asserts something narrower than before — only the literal `true` is now rejected, rather than every truthy value — which is the assertion this rule is asking for.

This rule is the opposite of `prefer-to-be-truthy` and `prefer-to-be-falsy`, which rewrite in the other direction. Enable this one or those two, never both.

Every `expect` source is recognized: globals, named and renamed imports, `require` destructuring, namespace imports, whole-module `require`, `import.meta.rstest`, Playwright integrations, and the `expect` a test callback receives through its [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). An `expect` from another assertion library, or a local variable that shadows `expect`, is not reported.

## Incorrect

```ts
test('reports a valid configuration as valid', () => {
  expect(isValidConfig(config)).toBeTruthy();
});
```

## Correct

```ts
test('reports a valid configuration as valid', () => {
  expect(isValidConfig(config)).toBe(true);
});
```

## Autofix

Renames the matcher to `toBe` and writes the matching literal into its argument list. Nothing else in the assertion changes: the expect root as written at the call site, `expect.soft`, the second `message` argument of `expect(actual, message)`, the modifier chain, the accessor's own quoting, and the surrounding line breaks and comments are all left as they are — a comment between the parentheses stays put, with the literal inserted before it.
