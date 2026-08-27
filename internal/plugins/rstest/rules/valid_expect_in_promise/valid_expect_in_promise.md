# valid-expect-in-promise

## Rule Details

Requires promise chains containing Rstest assertions to be returned or awaited. Otherwise the test can finish before the assertion runs or before its failure is reported.

Both Rstest `expect` and Chai `assert` calls are recognized. A promise chain may also be returned from the test callback.

## Incorrect

```ts
test('loads a user', () => {
  loadUser().then((user) => {
    expect(user.name).toBe('Ada');
  });
});
```

## Correct

```ts
test('loads a user', async () => {
  await loadUser().then((user) => {
    expect(user.name).toBe('Ada');
  });
});
```
