# no-control-regex

## Rule Details

Disallows control characters (U+0000 through U+001F) in regular expressions. Control characters are rarely intended in patterns and usually indicate a typo.

The rule flags:

- Unescaped raw characters in the U+0000–U+001F range
- `\xHH` escapes with `HH` in `00`–`1F`
- `\uHHHH` escapes with `HHHH` in `0000`–`001F`
- `\u{H...}` escapes (under the `u` or `v` flag) resolving to U+0000–U+001F

Symbolic control escapes such as `\t`, `\n`, `\r`, `\v`, `\f`, `\0`, and `\cX` are allowed.

Examples of **incorrect** code for this rule:

```javascript
var pattern1 = /\x00/;
var pattern2 = /\x0C/;
var pattern3 = /\x1F/;
var pattern4 = /\u000C/;
var pattern5 = /\u{C}/u;
var pattern6 = new RegExp('\x0C');
var pattern7 = new RegExp('\\x0C');
```

Examples of **correct** code for this rule:

```javascript
var pattern1 = /\x20/;
var pattern2 = /\u0020/;
var pattern3 = /\u{20}/u;
var pattern4 = /\t/;
var pattern5 = /\n/;
var pattern6 = new RegExp('\x20');
var pattern7 = new RegExp('\\t');
var pattern8 = new RegExp('\\n');
```

## Differences from ESLint

ESLint's implementation delegates pattern validation to `@eslint-community/regexpp` and wraps it in `try/catch`. On a regex-syntax error, regexpp aborts parsing, so any control characters appearing after the error point are never reported.

This implementation is a linear scanner without a full ES regex parser. On a **syntactically-invalid pattern** it keeps scanning past the error, which may surface control characters ESLint would have suppressed. For valid regex patterns the two implementations produce identical output.

Syntactically-invalid patterns are independently flagged by the [`no-invalid-regexp`](https://eslint.org/docs/latest/rules/no-invalid-regexp) rule, so running both rules together surfaces every relevant issue; only the rule attribution differs on malformed input.

## Original Documentation

- [no-control-regex](https://eslint.org/docs/latest/rules/no-control-regex)
