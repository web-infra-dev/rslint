# padding-around-describe-blocks

## Rule Details

Require a blank line before and after `describe` statements so suites are visually separated from surrounding setup, tests, and sibling suites. No trailing blank line is required when the suite is the last statement in its scope.

The rule classifies an expression statement by its first identifier, so calls and chains such as `describe.skip` and `describe.each` are included regardless of where `describe` was declared. Renamed aliases and namespace calls such as `rstest.describe` are not recognized. Type information is not required.

## Incorrect

```ts
const account = createAccount();
describe('account', () => {});
describe.skip('archived account', () => {});
```

## Correct

```ts
const account = createAccount();

describe('account', () => {});

describe.skip('archived account', () => {});
```

## Autofix

The rule inserts missing blank lines before and after `describe` statements.
