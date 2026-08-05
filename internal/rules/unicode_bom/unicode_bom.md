# unicode-bom

## Rule Details

The Unicode byte order mark (BOM, U+FEFF) marks whether code units are big endian or little endian. UTF-8 does not need one, because byte ordering does not matter when a character is a single byte, and UTF-8 dominates the web.

This rule controls whether a file begins with a BOM. Only the first position counts: a U+FEFF character anywhere else in the file is ordinary text and is left alone.

`rslint --fix` adds or removes the mark and rewrites the rest of the file unchanged.

## Options

This rule takes one string option.

### `"never"` (default)

A file must not begin with a byte order mark.

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

### `"always"`

A file must begin with a byte order mark.

```json
{ "unicode-bom": ["error", "always"] }
```

Example of **correct** code:

```javascript
// U+FEFF at the beginning
let abc;
```

Example of **incorrect** code:

```javascript
let abc;
```

The mark itself is invisible, so the comment stands in for it: the file's first three bytes are `EF BB BF`. A UTF-16 file, whose mark is `FF FE` or `FE FF`, counts as carrying one for the same reason.

## When Not To Use It

If you do not care about the presence of a byte order mark in your files, you can turn this rule off.

## Original Documentation

https://eslint.org/docs/latest/rules/unicode-bom
