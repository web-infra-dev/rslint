# no-conditional-expect

## Rule Details

Disallows assertions that only run on some execution paths. A test can pass without checking anything when its assertion is skipped by a condition.

The rule recognizes `expect` from globals, imports, the local test context, Browser Mode, and Playwright integrations.

## Incorrect

```ts
test('returns the user name', () => {
  if (user) {
    expect(user.name).toBe('Ada');
  }
});
```

## Correct

```ts
test('returns the user name', () => {
  expect(user).toBeDefined();
  expect(user.name).toBe('Ada');
});
```
