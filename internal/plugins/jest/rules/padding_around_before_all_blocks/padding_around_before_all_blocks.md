# padding-around-before-all-blocks

## Rule Details

Require a blank line before and after `beforeAll` statements. No trailing blank line is required when the hook is the last statement in its scope.

## Incorrect

```js
const database = createDatabase();
beforeAll(() => database.connect());
test('loads a user', loadUser);
```

## Correct

```js
const database = createDatabase();

beforeAll(() => database.connect());

test('loads a user', loadUser);
```

## Autofix

The rule inserts missing blank lines before and after `beforeAll` statements.

## Original Documentation

- [eslint-plugin-jest: padding-around-before-all-blocks](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/padding-around-before-all-blocks.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/padding-around-before-all-blocks.ts)
