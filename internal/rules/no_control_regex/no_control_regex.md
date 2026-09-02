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

## Original Documentation

- [ESLint: no-control-regex](https://eslint.org/docs/latest/rules/no-control-regex)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.1/lib/rules/no-control-regex.js)
