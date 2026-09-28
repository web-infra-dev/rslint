# no-error-equal

## Rule Details

Disallow using equality matchers on error types.

Jest's equality matchers compare errors mostly by their `message`. `toEqual` ignores the error type, so `new TypeError('x')` equals `new RangeError('x')`. `toStrictEqual` checks the type, but still ignores other properties such as `code` and `cause`. Passing an error instance to `toThrow` does not help either: `toThrow(new Error('x'))` compares only the message and passes even when the thrown value is not an `Error`.

Check the error type and the message separately instead. For a thrown error, pass the expected class to `toThrow` and assert the message in a second call. For an error value, use `toBeInstanceOf` and assert the properties you care about.

This rule reports `toEqual` and `toStrictEqual` assertions whose `expect()` subject has an `Error` type: the built-in `Error`, a built-in or user-defined class or interface that extends it, or a union whose members are all such types. It requires type information.

Examples of **incorrect** code for this rule:

```ts
expect(new AggregateError([], 'hello world')).toEqual(new Error('hello world'));

expect(new Error('hello world')).toStrictEqual('hello sunshine');
```

Examples of **correct** code for this rule:

```ts
expect(() => loadUser(id)).toThrow(NotFoundError);
expect(() => loadUser(id)).toThrow('user not found');

expect(error).toBeInstanceOf(AggregateError);
expect(error).toHaveProperty('message', 'hello world');
```

## Original Documentation

- [eslint-plugin-jest: no-error-equal](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/docs/rules/no-error-equal.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/no-error-equal.ts)
