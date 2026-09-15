# padding-around-after-all-blocks

## Rule Details

Require a blank line before and after `afterAll` statements. No trailing blank line is required when the hook is the last statement in its scope.

## Incorrect

```js
const database = createDatabase();
afterAll(() => database.close());
test('loads a user', loadUser);
```

## Correct

```js
const database = createDatabase();

afterAll(() => database.close());

test('loads a user', loadUser);
```

## Autofix

The rule inserts missing blank lines before and after `afterAll` statements.

## Original Documentation

- [eslint-plugin-jest: padding-around-after-all-blocks](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/padding-around-after-all-blocks.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/padding-around-after-all-blocks.ts)
