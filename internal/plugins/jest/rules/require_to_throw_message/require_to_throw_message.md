# require-to-throw-message

## Rule Details

Require a message for `toThrow()` (and its alias `toThrowError()`).

`toThrow()` checks that a function throws, as in `expect(() => a()).toThrow()`. Without a message, the assertion passes for any thrown error. Requiring a message ensures the intended error is thrown.

This rule reports when `toThrow()` or `toThrowError()` is used with no argument. Negated forms such as `.not.toThrow()` without a message are allowed.

Examples of **incorrect** code for this rule:

```js
test('all the things', async () => {
  expect(() => a()).toThrow();

  expect(() => a()).toThrowError();

  await expect(a()).rejects.toThrow();

  await expect(a()).rejects.toThrowError();
});
```

Examples of **correct** code for this rule:

```js
test('all the things', async () => {
  expect(() => a()).toThrow('a');

  expect(() => a()).toThrowError('a');

  await expect(a()).rejects.toThrow('a');

  await expect(a()).rejects.toThrowError('a');

  expect(() => a()).not.toThrow();
});
```

## Original Documentation

- [jest/require-to-throw-message](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/require-to-throw-message.md)
