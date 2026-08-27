# valid-expect

## Rule Details

Validates Rstest `expect` calls: they need the required arguments and a matcher, and asynchronous assertions must be awaited or returned so failures are not lost.

Rstest supports a message as the second `expect` argument and Chai-style property assertions; both are accepted by the rule.

## Incorrect

```ts
expect(user);

test('loads a user', async () => {
  expect(loadUser()).resolves.toMatchObject({ name: 'Ada' });
});
```

## Correct

```ts
expect(user).toMatchObject({ name: 'Ada' });

test('loads a user', async () => {
  await expect(loadUser()).resolves.toMatchObject({ name: 'Ada' });
});
```

## Options

```json
{
  "rstest/valid-expect": [
    "error",
    {
      "alwaysAwait": true,
      "asyncMatchers": ["toReject", "toResolve", "toSettle"],
      "maxArgs": 2
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `alwaysAwait` | `boolean` | `false` | Require async assertions to be awaited instead of allowing `return`. |
| `asyncMatchers` | `string[]` | `["toReject", "toResolve"]` | Matcher names treated as asynchronous. Setting it replaces the default list. |
| `minArgs` | `number` | `1` | Minimum number of arguments for `expect`. |
| `maxArgs` | `number` | `1` | Maximum number of arguments for `expect`. |

## Autofix

An async assertion that is neither awaited nor returned is fixed by adding `await`, and by marking the enclosing function `async` when it is not already. Argument-count and matcher problems are reported without a fix.
