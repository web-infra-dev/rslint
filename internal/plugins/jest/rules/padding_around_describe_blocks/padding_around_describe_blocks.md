# padding-around-describe-blocks

## Rule Details

Require a blank line before and after Jest suite statements, including `describe`, `fdescribe`, and `xdescribe`. No trailing blank line is required when the suite is the last statement in its scope.

## Incorrect

```js
const account = createAccount();
describe('account', () => {});
describe.skip('archived account', () => {});
```

## Correct

```js
const account = createAccount();

describe('account', () => {});

describe.skip('archived account', () => {});
```

## Autofix

The rule inserts missing blank lines before and after suite statements.

## Original Documentation

- [eslint-plugin-jest: padding-around-describe-blocks](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/padding-around-describe-blocks.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/padding-around-describe-blocks.ts)
