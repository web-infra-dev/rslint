# unicode-bom

## Rule Details

The Unicode byte order mark (BOM, U+FEFF) marks whether code units are big endian or little endian. UTF-8 does not need one, because byte ordering does not matter when a character is a single byte, and UTF-8 dominates the web.

This rule disallows a BOM at the very start of a file. Only the first position counts: a U+FEFF character anywhere else in the file is ordinary text and is left alone.

`rslint --fix` removes the mark and rewrites the rest of the file unchanged.

## Options

This rule takes one string option, `"never"`, which is also the default:

```json
{ "unicode-bom": ["error", "never"] }
```

Example of **correct** code:

```javascript
let abc;
```

Example of **incorrect** code:

```javascript
// U+FEFF at the beginning
let abc;
```

The mark itself is invisible, so the comment stands in for it: the file's first three bytes are `EF BB BF`. A UTF-16 file, whose mark is `FF FE` or `FE FF`, is incorrect for the same reason.

## Differences from ESLint

- ESLint's `"always"` option is not supported in rslint and is not planned. `"never"` is the only allowed option; anything else is rejected as an invalid configuration.

## When Not To Use It

If you do not care about the presence of a byte order mark in your files, you can turn this rule off.

## Original Documentation

https://eslint.org/docs/latest/rules/unicode-bom
