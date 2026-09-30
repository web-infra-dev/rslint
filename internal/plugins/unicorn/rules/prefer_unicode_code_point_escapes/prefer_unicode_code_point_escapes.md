# prefer-unicode-code-point-escapes

## Rule Details

Prefer Unicode code point escapes such as `\u{61}` over hexadecimal,
four-digit Unicode, and legacy octal escapes in strings and untagged templates.
Adjacent high and low surrogate escapes become a single code point escape.

Examples of **incorrect** code for this rule:

```javascript
const letter = '\x7A';
const symbol = '\uD83D\uDCA9';
const template = `\u2661${value}`;
const pattern = /\u0061/u;
```

Examples of **correct** code for this rule:

```javascript
const letter = '\u{7A}';
const symbol = '\u{1F4A9}';
const template = `\u{2661}${value}`;
const pattern = /\u{61}/u;
```

The rule has no options. Strings, untagged templates, and regular expressions
with a `u` or `v` flag are automatically fixable. Regular expression control
escapes such as `\cA` are also converted. Surrogate escapes inside regular
expression character classes are left unchanged.

For a regular expression without `u` or `v`, the rule offers a suggestion to
convert its escapes and add `u`. Adding Unicode mode can change matching
behavior, so this conversion is never an automatic fix. No suggestion is
offered if the converted pattern would be invalid.

Tagged template contents are ignored because tag functions can observe their
raw text. Escaped backslashes, common escapes such as `\n` and `\0`, and
existing code point escapes are left unchanged. The rule does not interpret
string arguments to `RegExp` as regular expression patterns.

## Differences from upstream

In JavaScript scripts, rslint reports legacy octal escapes such as `'\123'`
and non-octal decimal escapes such as `'\8'` as syntax errors, so this rule
cannot fix them. ESLint accepts them in non-strict scripts. Replace `'\123'`
with `'\u{53}'` and `'\8'` with `'8'` before linting with rslint.

## Original Documentation

- [eslint-plugin-unicorn: prefer-unicode-code-point-escapes](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/prefer-unicode-code-point-escapes.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/rules/prefer-unicode-code-point-escapes.js)
