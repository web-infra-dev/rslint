# padding-around-expect-groups

## Rule Details

Require a blank line around each consecutive group of `expect` statements. Adjacent assertions remain together without blank lines between them. Awaited `expect` statements are included.

## Incorrect

```js
const account = loadAccount();
expect(account.name).toBe('Ada');
expect(account.active).toBe(true);
saveAccount(account);
```

## Correct

```js
const account = loadAccount();

expect(account.name).toBe('Ada');
expect(account.active).toBe(true);

saveAccount(account);
```

## Autofix

The rule inserts missing blank lines at assertion-group boundaries.

## Original Documentation

- [eslint-plugin-jest: padding-around-expect-groups](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/padding-around-expect-groups.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/padding-around-expect-groups.ts)
