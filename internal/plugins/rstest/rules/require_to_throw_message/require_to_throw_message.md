# require-to-throw-message

## Rule Details

Requires `toThrow()` and its `toThrowError()` alias to specify the expected error. Without an argument, either matcher passes for any thrown error, so a test can continue passing after the code starts throwing for an unrelated reason. The argument may be a string, template literal, regular expression, error constructor, or any other expression accepted by Rstest's [`expect` API](https://rstest.rs/api/runtime-api/test-api/expect).

The rule checks the first called matcher in a call-style assertion chain, including `expect.soft(...)` and promise modifiers such as `rejects`. Negated forms such as `not.toThrow()` are exempt because they assert that no error is thrown. Chai-style `throw` and `throws` assertions are not checked. Although Rstest does not expose throw matchers on `expect.poll(...)` or `expect.element(...)`, the same syntax-only check applies if either form is written.

With type information, every Rstest `expect` source is recognized: globals, named and renamed imports, `require` destructuring, namespace imports, whole-module `require`, `import.meta.rstest`, Playwright integrations, and the `expect` supplied by a test's [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). The rule also runs without type information; in that mode it recognizes the global and direct bindings named `expect` from imports, `require` destructuring, `import.meta.rstest` destructuring, and a test callback's destructured TestContext. An `expect` imported from another assertion library or shadowed by a local value is ignored.

## Incorrect

```ts
test('rejects an invalid account', async () => {
  expect(() => loadAccount('missing')).toThrow();
  await expect(loadAccount('missing')).rejects.toThrowError();
});
```

## Correct

```ts
test('rejects an invalid account', async () => {
  expect(() => loadAccount('missing')).toThrow('Account not found');
  await expect(loadAccount('missing')).rejects.toThrowError(/not found/);
});
```
