# no-focused-tests

## Rule Details

Disallows focused tests and suites. Leaving `.only` in committed code prevents the rest of the suite from running.

The rule recognizes `.only` on Rstest tests and suites, including aliases, parameterized registrations, fixtures, and Playwright integrations.

## Incorrect

```ts
describe.only('user service', () => {
  test('creates a user', () => {});
});
```

## Correct

```ts
describe('user service', () => {
  test('creates a user', () => {});
});
```

## Suggestions

Offers a suggestion to remove `.only` when the focused modifier is unambiguous.
