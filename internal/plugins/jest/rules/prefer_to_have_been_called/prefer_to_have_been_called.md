# prefer-to-have-been-called

## Rule Details

In order to have a better failure message, `toHaveBeenCalled()` and
`not.toHaveBeenCalled()` should be used when asserting that a mock has or has
not been called.

This rule triggers a warning if `toHaveBeenCalledTimes` or `toBeCalledTimes` is
used to assert that a mock has or has not been called zero times.

Examples of **incorrect** code for this rule:

```js
expect(method).toHaveBeenCalledTimes(0);

expect(method).not.toHaveBeenCalledTimes(0);
```

Examples of **correct** code for this rule:

```js
expect(method).not.toHaveBeenCalled();

expect(method).toHaveBeenCalled();
```

## Original Documentation

- [jest/prefer-to-have-been-called](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-to-have-been-called.md)
