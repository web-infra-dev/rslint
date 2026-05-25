# no-empty-character-class

## Rule Details

Disallows empty character classes `[]` in regular expression literals. An empty character class in a regular expression does not match anything and is almost certainly a mistake. Note that `[^]` (a negated empty class) is allowed since it matches any character.

With the ES2024 `v` flag (unicodeSets), character classes can be nested. This rule also detects empty classes inside nested structures such as set subtraction (`--`) and intersection (`&&`).

This rule does not check `new RegExp()` constructor calls — only regex literals.

Examples of **incorrect** code for this rule:

```javascript
var foo = /^abc[]/;
var foo = /foo[]bar/;
var foo = /[]]/;
// v-flag (ES2024)
var foo = /[[]]/v;
var foo = /[a--[]]/v;
var foo = /[a&&[]]/v;
```

Examples of **correct** code for this rule:

```javascript
var foo = /^abc[a-zA-Z]/;
var foo = /[^]/;
var foo = /[\\[]/;
var foo = /\\[]/;
// v-flag (ES2024)
var foo = /[[^]]/v;
var foo = /[a--b]/v;
var foo = /[[a][b]]/v;
```

## Original Documentation

- [ESLint no-empty-character-class](https://eslint.org/docs/latest/rules/no-empty-character-class)
