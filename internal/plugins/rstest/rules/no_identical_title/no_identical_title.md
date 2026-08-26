# no-identical-title

## Rule Details

Requires unique static test and suite titles within the same scope. Distinct titles make failures and test reports unambiguous.

Test titles and suite titles are tracked separately. Dynamic titles and parameterized titles are ignored because their final values are created at runtime.

## Incorrect

```ts
test('creates a user', () => {});
test('creates a user', () => {});
```

## Correct

```ts
test('creates a user with a name', () => {});
test('rejects a user without a name', () => {});
```
