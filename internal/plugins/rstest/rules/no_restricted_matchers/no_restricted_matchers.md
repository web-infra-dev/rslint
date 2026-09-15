# no-restricted-matchers

## Rule Details

Disallow selected matcher or modifier chains in Rstest assertions. This lets a project steer tests away from matchers that are ambiguous, too broad, or inconsistent with its assertion conventions, with an optional message that names the preferred alternative.

The rule recognizes global and imported `expect`, renamed and namespace imports, CommonJS imports, `import.meta.rstest`, and the `expect` provided by a test's [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). It also recognizes `expect.soft`, `expect.poll`, `expect.element`, static `expect` APIs, and Chai-style language chains and property or call matchers. Local shadows and assertion functions from other frameworks are ignored. The rule does not require type information.

For a Chai chain containing multiple assertions, only the chain through the first matcher is checked. For example, `to.be.a` can restrict `expect(value).to.be.a('string').and.contain('x')`, while `contain` cannot. Dynamic computed members such as `expect(value)[matcher]()` are ignored because their names are not known statically.

## Incorrect

```ts
test('returns the complete account', () => {
  expect(loadAccount()).toEqual({ id: 1, name: 'Ada' });
});
```

## Correct

```ts
test('returns the complete account', () => {
  expect(loadAccount()).toStrictEqual({ id: 1, name: 'Ada' });
});
```

## Options

```json
{
  "rstest/no-restricted-matchers": [
    "error",
    {
      "toEqual": "Use toStrictEqual for object comparisons.",
      "resolves.not": null
    }
  ]
}
```

The option object maps a matcher or modifier chain to a custom diagnostic message, or to `null` for the default message. Ordinary chains require an exact match. A single `not`, `resolves`, or `rejects` entry matches any chain beginning with that modifier, and a chain ending in `.not` also matches by prefix.

When several entries match, the chain with more segments wins; chains of equal length are ordered alphabetically. Only one diagnostic is reported for each assertion.

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| _(matcher or modifier chain)_ | `string \| null` | `{}` | Custom diagnostic message, or `null` for the default message. |
