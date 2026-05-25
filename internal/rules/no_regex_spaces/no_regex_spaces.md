# no-regex-spaces

## Rule Details

Disallow multiple spaces in regular expressions. Two or more consecutive space
characters are hard to count by eye; an explicit `{n}` quantifier expresses the
same pattern unambiguously. The rule applies to both regex literals and the
`RegExp` / `new RegExp` constructors (skipped when `RegExp` is shadowed in the
enclosing scope or when the flags argument cannot be statically determined).
Consecutive spaces inside character classes (`[...]`) are intentionally allowed
and not reported.

Examples of **incorrect** code for this rule:

```javascript
var re = /foo   bar/;
var re = new RegExp('foo   bar');
```

Examples of **correct** code for this rule:

```javascript
var re = /foo {3}bar/;
var re = new RegExp('foo {3}bar');
var re = /[  ]/;
```

## Autofix

When the parsed pattern and the raw source text agree (typically any regex
literal, and `RegExp` / `new RegExp` calls whose pattern string contains no
escape sequences), the rule rewrites `  ` into ` {n}`. When the pattern
contains escape sequences that differ from the raw source (e.g.
`new RegExp('\\d  ')`), the rule reports but does not autofix — the index into
the parsed pattern would not map cleanly back to source positions.

## Original Documentation

- ESLint rule: [https://eslint.org/docs/latest/rules/no-regex-spaces](https://eslint.org/docs/latest/rules/no-regex-spaces)
- Source code: [https://github.com/eslint/eslint/blob/main/lib/rules/no-regex-spaces.js](https://github.com/eslint/eslint/blob/main/lib/rules/no-regex-spaces.js)
