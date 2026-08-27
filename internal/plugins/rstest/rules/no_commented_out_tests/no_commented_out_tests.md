# no-commented-out-tests

## Rule Details

Disallows tests and suites hidden in comments. Commented-out tests are easy to miss in review and can remain stale indefinitely.

Use `test.todo` when a test should remain visible but is not ready to run. The rule recognizes commented `test`, `it`, and `describe` registrations, including modifiers and parameterized forms.

## Incorrect

```ts
// test('rejects an invalid email', () => {
//   expect(validateEmail('invalid')).toBe(false);
// });
```

## Correct

```ts
test.todo('rejects an invalid email');
```
