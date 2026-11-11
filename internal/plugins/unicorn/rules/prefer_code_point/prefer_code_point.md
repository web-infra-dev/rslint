# prefer-code-point

## Rule Details

[`String#codePointAt()`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/codePointAt) and [`String.fromCodePoint()`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/fromCodePoint) handle the full Unicode range, including characters outside the Basic Multilingual Plane that JavaScript represents with surrogate pairs. `String#charCodeAt()` and `String.fromCharCode()` only operate on individual UTF-16 code units, so they truncate emoji and other high code points.

Examples of **incorrect** code for this rule:

```javascript
const unicorn = '🦄'.charCodeAt(0).toString(16);
const first = String.fromCharCode(0x1f984);
```

Examples of **correct** code for this rule:

```javascript
const unicorn = '🦄'.codePointAt(0).toString(16);
const first = String.fromCodePoint(0x1f984);
```

## Suggestions

The rule offers a suggestion that renames `charCodeAt` to `codePointAt` only when the result is not consumed as a number. In string hashing or digit parsing such as `string.charCodeAt(index) - 48`, swapping in `codePointAt` would change the computed value: a code point outside the BMP occupies two code units, and iterating by `string.length` would still visit the trailing surrogate separately. The diagnostic is still reported there, but the fix is left for manual review.

```javascript
// Reported without a suggestion: a correct code-point hash must iterate by code point.
let hash = 0;
for (const character of string) {
	hash = (((hash * 31) | 0) + character.codePointAt(0)) | 0;
}
```

`String.fromCharCode` is always reported with a suggestion because the call shape does not change; callers that intentionally build strings from UTF-16 code units can disable the rule for that line.

## Original Documentation

- [eslint-plugin-unicorn: prefer-code-point](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-code-point.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-code-point.js)
