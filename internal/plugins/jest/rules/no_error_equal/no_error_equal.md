# no-error-equal

## Rule Details

Disallow using equality matchers on error types.

When comparing errors, `toEqual` and `toStrictEqual` only compare the `message` properties, so a test can pass even if the errors are of different types. Use `toThrow` instead, which checks the error type along with its message.

This rule reports `toEqual` and `toStrictEqual` assertions whose `expect()` subject has an `Error` type: the built-in `Error`, a built-in or user-defined class or interface that extends it, or a union whose members are all such types. It requires type information.

Examples of **incorrect** code for this rule:

```ts
expect(new AggregateError([], expect.any(String))).toEqual(
  new Error(expect.any(String)),
);

expect(new Error('hello world')).toStrictEqual('hello sunshine');
```

Examples of **correct** code for this rule:

```ts
expect(() => {
  throw new AggregateError([], expect.any(String));
}).toThrow(new Error(expect.any(String)));

expect(() => {
  throw new Error('hello world');
}).toThrow('hello sunshine');
```

## Original Documentation

- [eslint-plugin-jest: no-error-equal](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/docs/rules/no-error-equal.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/no-error-equal.ts)
