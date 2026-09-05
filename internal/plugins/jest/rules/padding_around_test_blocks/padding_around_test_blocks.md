# padding-around-test-blocks

## Rule Details

Require a blank line before and after Jest test statements, including `test`, `it`, `fit`, `xit`, and `xtest`. No trailing blank line is required when the test is the last statement in its scope.

## Incorrect

```js
const account = createAccount();
test('saves the account', saveAccount);
it('loads the account', loadAccount);
```

## Correct

```js
const account = createAccount();

test('saves the account', saveAccount);

it('loads the account', loadAccount);
```

## Autofix

The rule inserts missing blank lines before and after test statements.

## Original Documentation

- [eslint-plugin-jest: padding-around-test-blocks](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/padding-around-test-blocks.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/padding-around-test-blocks.ts)
