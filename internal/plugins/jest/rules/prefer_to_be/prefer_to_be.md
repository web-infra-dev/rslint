# prefer-to-be

## Rule Details

This rule promotes the use of the most appropriate equality matcher in `expect`
assertions, which results in more idiomatic tests and clearer failure output:

- For primitive literals (numbers, strings, booleans), prefer `toBe()` over
  `toEqual()` or `toStrictEqual()`. The matchers behave identically here, but
  `toBe()` reads more naturally.
- For `null`, `undefined`, and `NaN`, prefer the dedicated matchers
  `toBeNull()`, `toBeUndefined()` / `toBeDefined()`, and `toBeNaN()`. They
  produce more descriptive error messages than `toBe()`, `toEqual()`, or
  `toStrictEqual()`.

The rule is reported through the following messages:

| Message ID         | Description                              |
| ------------------ | ---------------------------------------- |
| `useToBe`          | Use `toBe` when expecting primitive literals |
| `useToBeUndefined` | Use `toBeUndefined` instead              |
| `useToBeDefined`   | Use `toBeDefined` instead                |
| `useToBeNull`      | Use `toBeNull` instead                   |
| `useToBeNaN`       | Use `toBeNaN` instead                    |

Examples of **incorrect** code for this rule:

```js
// Use `toBe` for primitive literals
expect(value).not.toEqual(5);
expect(getMessage()).toStrictEqual('hello world');
expect(loadMessage()).resolves.toEqual('hello world');

// Use the dedicated matchers for `null` / `undefined` / `NaN`
expect(value).not.toBe(undefined);
expect(getMessage()).toBe(null);
expect(countMessages()).resolves.not.toBe(NaN);
```

Examples of **correct** code for this rule:

```js
expect(value).not.toBe(5);
expect(getMessage()).toBe('hello world');
expect(loadMessage()).resolves.toBe('hello world');
expect(didError).not.toBe(true);

expect(value).toBeDefined();
expect(getMessage()).toBeNull();
expect(countMessages()).resolves.not.toBeNaN();

// Object/array literals still use `toStrictEqual`
expect(catchError()).toStrictEqual({ message: 'oh noes!' });
expect(catchError()).toStrictEqual({ message: undefined });
```

## Original Documentation

- [jest/prefer-to-be](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-to-be.md)
