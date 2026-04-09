# jest/prefer-strict-equal

## Rule Details

Prefer `toStrictEqual()` over `toEqual()` on `expect()`. It is common to expect objects to not only have identical values but also to have identical keys. A stricter equality will catch cases where two objects do not have identical keys.

Examples of **incorrect** code for this rule:

```javascript
expect({ a: 'a', b: undefined }).toEqual({ a: 'a' });
```

Examples of **correct** code for this rule:

```javascript
expect({ a: 'a', b: undefined }).toStrictEqual({ a: 'a' });
```

## Original Documentation

- [jest/prefer-strict-equal](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-strict-equal.md)
