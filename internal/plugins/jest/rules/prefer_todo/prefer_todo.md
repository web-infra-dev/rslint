# jest/prefer-todo

## Rule Details

Prefer test.todo('…') (or the same on it / describe-style test APIs) when a test is only a title, has an empty implementation (() => {}), or is test.skip with an empty body. That makes unfinished work explicit in Jest’s reporting instead of looking like a passing or silently skipped test.
****
Examples of **incorrect** code for this rule:

```js
test('i need to write this test'); // Unimplemented test case
test('i need to write this test', () => {}); // Empty test case body
test.skip('i need to write this test', () => {}); // Empty test case body
```

Example of **correct** code for this rule:

```js
test.todo('i need to write this test');
```

## Original Documentation

- [jest/prefer-todo](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-todo.md)
