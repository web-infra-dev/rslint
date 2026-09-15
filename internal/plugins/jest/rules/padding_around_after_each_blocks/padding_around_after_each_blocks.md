# padding-around-after-each-blocks

## Rule Details

Require a blank line before and after `afterEach` statements. No trailing blank line is required when the hook is the last statement in its scope.

## Incorrect

```js
const database = createDatabase();
afterEach(() => database.reset());
test('loads a user', loadUser);
```

## Correct

```js
const database = createDatabase();

afterEach(() => database.reset());

test('loads a user', loadUser);
```

## Autofix

The rule inserts missing blank lines before and after `afterEach` statements.

## Original Documentation

- [eslint-plugin-jest: padding-around-after-each-blocks](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/padding-around-after-each-blocks.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/padding-around-after-each-blocks.ts)
