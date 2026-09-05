# padding-around-test-blocks

## Rule Details

Require a blank line before and after `test` and `it` statements so individual test registrations are visually separated. No trailing blank line is required when the test is the last statement in its scope.

The rule classifies an expression statement by its first identifier, so calls and chains such as `test.skip`, `test.each`, and fixture-based calls beginning with `test` are included regardless of where `test` or `it` was declared. Renamed aliases and namespace calls such as `rstest.test` are not recognized. Type information is not required.

## Incorrect

```ts
const account = createAccount();
test('saves the account', saveAccount);
it('loads the account', loadAccount);
```

## Correct

```ts
const account = createAccount();

test('saves the account', saveAccount);

it('loads the account', loadAccount);
```

## Autofix

The rule inserts missing blank lines before and after `test` and `it` statements.
