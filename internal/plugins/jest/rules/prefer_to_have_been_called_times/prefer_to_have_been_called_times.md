# prefer-to-have-been-called-times

## Rule Details

In order to have a better failure message, `toHaveBeenCalledTimes()` should be
used instead of directly checking the length of `mock.calls`.

This rule triggers a warning if `toHaveLength` is used to assert the number of
times a mock has been called.

This rule is often used together with
[`prefer-to-have-length`](./prefer_to_have_length.md).

Examples of **incorrect** code for this rule:

```js
expect(someFunction.mock.calls).toHaveLength(1);
expect(someFunction.mock.calls).toHaveLength(0);

expect(someFunction.mock.calls).not.toHaveLength(1);
```

Examples of **correct** code for this rule:

```js
expect(someFunction).toHaveBeenCalledTimes(1);
expect(someFunction).toHaveBeenCalledTimes(0);

expect(someFunction).not.toHaveBeenCalledTimes(0);

expect(uncalledFunction).not.toBeCalled();

expect(method.mock.calls[0][0]).toStrictEqual(value);
```

## Original Documentation

- [jest/prefer-to-have-been-called-times](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-to-have-been-called-times.md)
