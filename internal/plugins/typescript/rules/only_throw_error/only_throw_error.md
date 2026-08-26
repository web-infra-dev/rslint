# only-throw-error

## Rule Details

Disallow throwing non-Error values as exceptions. It is considered good practice to only `throw` `Error` objects, because they automatically capture a stack trace which can be used to debug the error. Throwing non-Error values such as strings, numbers, or `undefined` does not provide this benefit and makes debugging harder.

Examples of **incorrect** code for this rule:

```typescript
throw 'error';
throw 0;
throw undefined;
throw { message: 'error' };
```

Examples of **correct** code for this rule:

```typescript
throw new Error('error');
throw new RangeError('error');

class CustomError extends Error {}
throw new CustomError('error');
```

## Original Documentation

- [typescript-eslint: only-throw-error](https://typescript-eslint.io/rules/only-throw-error)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/only-throw-error.ts)
