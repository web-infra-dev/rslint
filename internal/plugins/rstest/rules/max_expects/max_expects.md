# max-expects

## Rule Details

Limits completed assertion chains in each Rstest test or hook callback. A smaller test is usually easier to understand because it verifies one behavior at a time.

Each `expect(...)`, `expect.soft(...)`, `expect.poll(...)`, or `expect.element(...)` assertion counts once. Static helpers such as `expect.any()` do not count.

## Incorrect

```ts
test('returns an active user', () => {
  expect(user.id).toBe(1);
  expect(user.name).toBe('Ada');
  expect(user.active).toBe(true);
  expect(user.role).toBe('admin');
  expect(user.email).toBe('ada@example.com');
  expect(user.createdAt).toBeInstanceOf(Date);
});
```

## Correct

```ts
test('returns the user identity', () => {
  expect(user).toMatchObject({ id: 1, name: 'Ada' });
});

test('returns an active user', () => {
  expect(user.active).toBe(true);
});
```

## Options

```json
{
  "rstest/max-expects": [
    "error",
    {
      "max": 2
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `max` | `integer` | `5` | Maximum assertion calls allowed in one test or hook callback. |
