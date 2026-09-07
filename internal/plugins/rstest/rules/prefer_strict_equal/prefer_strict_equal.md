# prefer-strict-equal

## Rule Details

Requires `toStrictEqual()` for structural equality assertions. Unlike `toEqual()`, strict equality checks that objects have the same keys and does not ignore properties whose value is `undefined`, preventing incomplete expected values from passing unnoticed.

The rule checks call-style `toEqual()` matchers on global, imported, required, namespace and `import.meta.rstest` APIs, as well as `expect` obtained from a test's [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). It also recognizes `expect.soft()`, `expect.poll()` and the syntactic `expect.element()` form. Only the first matcher in an assertion chain is checked; Chai-style matchers such as `equal()` and property assertions are left unchanged. Static `expect` APIs, locally shadowed or foreign `expect` functions, and matcher names computed at runtime are ignored.

## Incorrect

```ts
test('returns the complete profile', () => {
  expect(loadProfile()).toEqual({ name: 'Ada' });
});
```

## Correct

```ts
test('returns the complete profile', () => {
  expect(loadProfile()).toStrictEqual({ name: 'Ada' });
});
```

## Suggestions

Replaces the `toEqual` matcher accessor with `toStrictEqual`. Dot, string-literal and template-literal accessors keep their original syntax, and arguments, comments and the spelling of the `expect` root are unchanged. If an accessor cannot be rewritten safely, the rule reports it without a suggestion.
