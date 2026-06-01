# prefer-called-with

## Rule Details

The `toHaveBeenCalled()` and `toBeCalled()` matchers assert that a mock function
has been called one or more times, without checking the arguments passed. The
assertion is stronger when arguments are also validated using
`toHaveBeenCalledWith()` or `toBeCalledWith()`. When some arguments are difficult
to check, using generic matchers such as `expect.anything()` at least enforces the
number and position of arguments.

Examples of **incorrect** code for this rule:

```js
expect(someFunction).toBeCalled();

expect(someFunction).toHaveBeenCalled();
```

Examples of **correct** code for this rule:

```js
expect(noArgsFunction).toHaveBeenCalledWith();

expect(roughArgsFunction).toHaveBeenCalledWith(
  expect.anything(),
  expect.any(Date),
);

expect(anyArgsFunction).toHaveBeenCalledTimes(1);

expect(uncalledFunction).not.toHaveBeenCalled();
```

## Original Documentation

- [jest/prefer-called-with](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-called-with.md)
