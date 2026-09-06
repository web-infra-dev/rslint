# prefer-equality-matcher

## Rule Details

Requires strict equality checks inside `expect` to use an equality matcher directly. `expect(actual === expected).toBe(true)` is harder to read and produces less useful failures than `expect(actual).toBe(expected)`. The rule handles both `===` and `!==`, boolean expectations, an existing `.not`, and the `resolves` and `rejects` modifiers.

The rule recognizes `toBe`, `toEqual`, and `toStrictEqual` when they are the first called matcher in an assertion. It recognizes Rstest globals, imports and aliases, namespace access, `import.meta.rstest`, and the `expect` supplied by a test's [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). It also preserves assertion factories such as `expect.soft`, the second message argument of `expect(actual, message)`, matcher argument trivia, and trailing commas.

Loose comparisons (`==` and `!=`), Chai property matchers, later matchers in a Chai chain, dynamic computed matcher names, locally shadowed names, and `expect` imported from another test framework are not reported. `expect.poll(fn)` and `expect.element(locator)` are only reported when the value passed to that factory is itself a strict equality comparison.

## Incorrect

```ts
test('matches the saved profile', () => {
  expect(savedProfile === expectedProfile).toBe(true);
  expect(savedStatus !== 'archived').not.toEqual(false);
});
```

## Correct

```ts
test('matches the saved profile', () => {
  expect(savedProfile).toBe(expectedProfile);
  expect(savedStatus).not.toEqual('archived');
});
```

## Suggestions

Each report offers three alternatives using `toBe`, `toEqual`, and `toStrictEqual`. The suggestion moves the left side of the comparison into `expect`, moves the right side into the matcher, and adjusts `.not` so the assertion keeps the same meaning. `resolves` or `rejects`, the written expect root, a second message argument, accessor quoting, surrounding trivia, and a trailing matcher-argument comma remain unchanged.

If comments or unusual accessor syntax inside an edited span cannot be preserved safely, the rule reports the assertion without suggestions.
