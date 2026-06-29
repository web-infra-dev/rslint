# prefer-jest-mocked

## Rule Details

Prefer `jest.mocked()` over type assertions such as `fn as jest.Mock` when
working with mocked functions in Jest. The `jest.mocked()` helper preserves
correct mock typings without manual casts, which improves type safety and keeps
mock setup easier to read.

This rule reports type assertions and angle-bracket casts to Jest mock types,
including chained assertions like `as unknown as jest.Mock`. It is fixable.

Restricted types:

- `jest.Mock`
- `jest.MockedFunction`
- `jest.MockedClass`
- `jest.MockedObject`

Examples of **incorrect** code for this rule:

```typescript
(foo as jest.Mock).mockReturnValue(1);
const mock = (foo as jest.Mock).mockReturnValue(1);
(foo as unknown as jest.Mock).mockReturnValue(1);
(Obj.foo as jest.Mock).mockReturnValue(1);
([].foo as jest.Mock).mockReturnValue(1);
```

Examples of **correct** code for this rule:

```typescript
jest.mocked(foo).mockReturnValue(1);
const mock = jest.mocked(foo).mockReturnValue(1);
jest.mocked(Obj.foo).mockReturnValue(1);
jest.mocked([].foo).mockReturnValue(1);
```

## Original Documentation

- [jest/prefer-jest-mocked](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-jest-mocked.md)
