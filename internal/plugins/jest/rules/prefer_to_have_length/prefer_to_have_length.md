# prefer-to-have-length

## Rule Details

Prefer `toHaveLength()` when asserting the length of a value. It produces clearer
failure output than comparing `.length` with `toBe()`, `toEqual()`, or
`toStrictEqual()`.

Examples of **incorrect** code for this rule:

```js
expect(files.length).toBe(1);

expect(files.length).toEqual(1);

expect(files.length).toStrictEqual(1);
```

Examples of **correct** code for this rule:

```js
expect(files).toHaveLength(1);
```

## Original Documentation

- [eslint-plugin-jest: prefer-to-have-length](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/prefer-to-have-length.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/prefer-to-have-length.ts)
