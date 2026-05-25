# prefer-strict-equal

## Rule Details

Prefer `toStrictEqual()` over `toEqual()`. While `toEqual` recursively checks every field of an object or array, it ignores `undefined` properties. This can lead to unexpected test passes and hide bugs. `toStrictEqual()` provides a stricter equality check that does not ignore `undefined` properties, making your tests more robust.

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
