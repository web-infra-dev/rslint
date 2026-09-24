# jest/no-untyped-mock-factory

Require a type argument on `jest.mock` and `jest.doMock`, or an explicit return type on their inline factories, so TypeScript can check the mocked module shape.

## Rule Details

This rule checks calls with exactly two arguments. Calls with no factory, a third argument such as `{ virtual: true }`, an explicit type argument (including `any`), or an inline factory return annotation are exempt. A separately declared factory still needs a type argument on the mock call.

Enable this rule only for TypeScript files: its fixes insert TypeScript syntax. It does not require type information and does not check whether the annotation matches the actual module.

## Incorrect

```ts
jest.mock('./user-service', () => ({
  fetchUser: jest.fn(),
}));
```

## Correct

```ts
jest.mock<typeof import('./user-service')>('./user-service', () => ({
  fetchUser: jest.fn(),
}));

jest.doMock('./user-service', (): typeof import('./user-service') => ({
  fetchUser: jest.fn(),
}));
```

## Autofix

Adds `<typeof import('./module')>` when the first argument is a quoted string, preserving its quotes and escapes. Other module expressions are reported without a fix.

## Differences from ESLint

For parenthesized calls such as `(jest.mock)(path, factory)`, the type argument is inserted after the closing parenthesis, directly on the call.

For optional calls such as `jest.mock?.('./user-service', factory)`, the type argument is inserted after `?.` directly on the optional call.

## Original Documentation

- [eslint-plugin-jest: no-untyped-mock-factory](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/docs/rules/no-untyped-mock-factory.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/no-untyped-mock-factory.ts)
