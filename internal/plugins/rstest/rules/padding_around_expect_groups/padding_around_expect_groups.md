# padding-around-expect-groups

## Rule Details

Require a blank line around each consecutive group of `expect` statements so assertions are visually separated from test setup and subsequent actions. Adjacent assertions remain together without blank lines between them, and no trailing blank line is required when the group ends its scope.

The rule classifies an expression statement by its first identifier, so direct and awaited assertions beginning with `expect` are included regardless of where that name was declared. Renamed aliases, `expectTypeOf`, and namespace calls such as `rstest.expect` are not recognized. Type information is not required.

## Incorrect

```ts
test('loads an active account', async () => {
  const account = await loadAccount();
  expect(account.name).toBe('Ada');
  expect(account.active).toBe(true);
  saveAccount(account);
});
```

## Correct

```ts
test('loads an active account', async () => {
  const account = await loadAccount();

  expect(account.name).toBe('Ada');
  expect(account.active).toBe(true);

  saveAccount(account);
});
```

## Autofix

The rule inserts missing blank lines at assertion-group boundaries.
