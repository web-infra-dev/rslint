# unicode-bom

## Rule Details

The Unicode byte order mark (BOM, U+FEFF) marks whether code units are big endian or little endian. UTF-8 does not need one, because byte ordering does not matter when a character is a single byte, and UTF-8 dominates the web.

This rule controls whether a file begins with a BOM. Only the first position counts: a U+FEFF character anywhere else in the file is ordinary text and is left alone.

The mark lives in a file's bytes rather than in its text, so this rule reads the file itself. It runs wherever rslint reads files — the CLI and the API — and `rslint --fix` adds or removes the mark, rewriting the rest of the file unchanged.

Editors work from decoded text: VS Code turns a leading mark into the document's encoding, shown in the status bar as "UTF-8 with BOM". The language server therefore leaves this rule to the CLI and the API, where the file's own bytes are in reach.

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

- [ESLint: unicode-bom](https://eslint.org/docs/latest/rules/unicode-bom)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/unicode-bom.js)
