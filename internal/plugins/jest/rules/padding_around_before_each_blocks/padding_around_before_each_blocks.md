# padding-around-before-each-blocks

## Rule Details

Require a blank line before and after `beforeEach` statements. No trailing blank line is required when the hook is the last statement in its scope.

## Incorrect

```js
const database = createDatabase();
beforeEach(() => database.reset());
test('loads a user', loadUser);
```

## Correct

```js
const database = createDatabase();

beforeEach(() => database.reset());

test('loads a user', loadUser);
```

## Autofix

The rule inserts missing blank lines before and after `beforeEach` statements.

## Original Documentation

- [eslint-plugin-jest: padding-around-before-each-blocks](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/padding-around-before-each-blocks.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/padding-around-before-each-blocks.ts)
