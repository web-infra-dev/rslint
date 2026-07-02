# prefer-equality-matcher

## Rule Details

Suggest using the built-in equality matchers.

Jest has built-in matchers for expecting equality, which allow for more readable
tests and error messages if an expectation fails.

This rule checks for _strict_ equality checks (`===` and `!==`) in tests that
could be replaced with one of the following built-in equality matchers:

- `toBe`
- `toEqual`
- `toStrictEqual`

Fixes are provided as editor suggestions (one suggestion per matcher above).

Loose equality (`==` / `!=`) is not reported.

Examples of **incorrect** code for this rule:

```js
expect(x === 5).toBe(true);
expect(name === 'Carl').not.toEqual(true);
expect(myObj !== thatObj).toStrictEqual(true);
```

Examples of **correct** code for this rule:

```js
expect(x).toBe(5);
expect(name).not.toEqual('Carl');
expect(myObj).not.toStrictEqual(thatObj);
```

## Original Documentation

- [jest/prefer-equality-matcher](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-equality-matcher.md)
