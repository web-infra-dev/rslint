# no-standalone-expect

## Rule Details

Disallows executable assertions outside a test callback. Assertions at module scope or directly in a suite run while the test file is registered, not as part of a test case.

Assertions inside helper functions are allowed because a test can call them. Static helpers such as `expect.any()` and `expect.extend()` are also allowed at module scope.

## Incorrect

```ts
describe('user service', () => {
  expect(createUser({ name: 'Ada' }).name).toBe('Ada');
});
```

## Correct

```ts
describe('user service', () => {
  test('creates a user', () => {
    expect(createUser({ name: 'Ada' }).name).toBe('Ada');
  });
});
```

## Options

```json
{
  "rstest/no-standalone-expect": [
    "error",
    { "additionalTestBlockFunctions": ["scenario"] }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `additionalTestBlockFunctions` | `string[]` | `[]` | Additional functions whose callbacks are treated as test blocks. |
