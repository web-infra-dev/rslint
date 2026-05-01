# jest/prefer-to-contain

## Rule Details

In order to have a better failure message, `toContain()` should be used upon asserting expectations on an array containing an object.

This rule triggers a warning if `toBe()`, `toEqual()` or `toStrictEqual()` is used to assert object inclusion in an array.

Examples of **incorrect** code for this rule:

```js
expect(a.includes(b)).toBe(true);
expect(a.includes(b)).not.toBe(true);
expect(a.includes(b)).toBe(false);
expect(a.includes(b)).toEqual(true);
expect(a.includes(b)).toStrictEqual(true);
```

Examples of **correct** code for this rule:

```js
expect(a).toContain(b);
expect(a).not.toContain(b);
```

## Original Documentation

- [jest/prefer-to-contain](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-to-contain.md)
