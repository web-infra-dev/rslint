# valid-expect-with-promise

## Rule Details

Requires assertions on Promise values to use `resolves` or `rejects`, and reports those modifiers when the subject is not a Promise. This prevents an assertion from inspecting the Promise object itself and catches promise modifiers that would fail at runtime. The rule uses TypeScript type information and distinguishes Promise instances from Promise constructors.

Ordinary `expect` and `expect.soft` assertions are checked, including Chai-style chains and promise modifiers between Chai assertions. Global expect calls, imports and aliases from Rstest modules, `import.meta.rstest`, and test-context expect calls are recognized. `expect.poll` and `expect.element` are excluded because they have their own asynchronous assertion semantics.

Rstest calls a function passed to `expect(...).rejects` without arguments. The rule therefore checks the return type selected for that zero-argument call. A function returning a Promise is valid with `rejects`; a Promise-returning function used with `resolves` is still a function subject and is reported. See the [Rstest expect API](https://rstest.rs/api/runtime-api/test-api/expect) for Promise, soft, polling, and browser-element assertions.

## Incorrect

```ts
test('loads the current user', async () => {
  expect(loadCurrentUser()).toMatchObject({ name: 'Ada' });

  await expect({ name: 'Ada' }).resolves.toMatchObject({ name: 'Ada' });
});

test('rejects invalid credentials', async () => {
  await expect(() => validateCredentials()).rejects.toThrow();
});
```

## Correct

```ts
test('loads the current user', async () => {
  await expect(loadCurrentUser()).resolves.toMatchObject({ name: 'Ada' });

  expect({ name: 'Ada' }).toMatchObject({ name: 'Ada' });
});

test('rejects invalid credentials', async () => {
  await expect(() => validateCredentialsAsync()).rejects.toThrow();
});
```

## Options

```json
{
  "rstest/valid-expect-with-promise": [
    "error",
    {
      "checkThenables": true
    }
  ]
}
```

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| `checkThenables` | `boolean` | `false` | Treat values whose `then` method accepts both fulfillment and rejection callbacks as Promise values. |
