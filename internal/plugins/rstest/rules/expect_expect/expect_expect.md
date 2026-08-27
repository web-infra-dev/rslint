# expect-expect

## Rule Details

Requires every Rstest `test` or `it` callback to contain an assertion. This prevents tests from passing after exercising code without verifying its result.

By default, both `expect` and Chai's `assert` count as assertions. `test.todo` and `it.todo` are exempt because they intentionally have no callback; globals, imports, aliases, parameterized tests, fixtures, and Playwright integrations are recognized.

## Incorrect

```ts
test('creates a user', async () => {
  await createUser({ name: 'Ada' });
});
```

## Correct

```ts
test('creates a user', async () => {
  const user = await createUser({ name: 'Ada' });

  expect(user.name).toBe('Ada');
});
```

## Options

```json
{
  "rstest/expect-expect": [
    "error",
    {
      "assertFunctionNames": ["expect", "assert", "assertUser"],
      "additionalTestBlockFunctions": ["scenario"]
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `assertFunctionNames` | `string[]` | `["expect", "assert"]` | Function names or patterns treated as assertions. Setting it replaces the default list. |
| `additionalTestBlockFunctions` | `string[]` | `[]` | Additional functions whose callbacks are treated as test blocks. |
