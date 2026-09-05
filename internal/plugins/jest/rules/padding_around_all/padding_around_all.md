# padding-around-all

## Rule Details

Require blank lines around Jest lifecycle hooks, suites, tests, and assertion groups. This rule combines all seven `padding-around-*` rules in one configuration entry. Consecutive `expect` statements form one group and do not require blank lines between them.

## Incorrect

```js
const database = createDatabase();
beforeAll(() => database.connect());
test('loads a user', () => {
  const user = loadUser();
  expect(user.name).toBe('Ada');
});
```

## Correct

```js
const database = createDatabase();

beforeAll(() => database.connect());

test('loads a user', () => {
  const user = loadUser();

  expect(user.name).toBe('Ada');
});
```

## Autofix

The rule inserts missing blank lines around matching statements and assertion groups.

## Original Documentation

- [eslint-plugin-jest: padding-around-all](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/padding-around-all.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/padding-around-all.ts)
