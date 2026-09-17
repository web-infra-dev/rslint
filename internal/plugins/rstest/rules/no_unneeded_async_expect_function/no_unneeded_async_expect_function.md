# no-unneeded-async-expect-function

## Rule Details

Disallow passing an async function wrapper to an Rstest `.resolves` assertion. Rstest requires `.resolves` to receive a Promise rather than a function. Pass a Promise directly to `expect` or `expect.soft`; when the inner call does not return a Promise, use an assertion that matches the value's actual behavior instead.

The rule checks `.resolves` chains whose first argument is an async arrow or function expression containing only `await someCall()`. It recognizes Rstest globals, imports and aliases from Rstest modules, namespace and CommonJS access, `import.meta.rstest`, and `expect` supplied by a test context. Jest-style and Chai-style matcher chains after `.resolves` are both recognized.

The rule does not check `.rejects`, because an async wrapper can be necessary to turn a synchronous exception into a rejected Promise. It also excludes `expect.poll` and `expect.element`. Wrappers containing comments are still reported. No type information is required.

The rule does not provide an autofix because source-only analysis cannot prove that the inner call returns a Promise or that moving the call before `expect` preserves evaluation order.

## Incorrect

```ts
test('resolves a user', async () => {
  await expect(async () => {
    await loadUser();
  }).resolves.toEqual({ name: 'Ada' });
});
```

## Correct

```ts
test('resolves a user', async () => {
  await expect(loadUser()).resolves.toEqual({ name: 'Ada' });
});
```
