# no-empty-character-class

## Rule Details

Disallows empty character classes `[]` in regular expression literals. An empty character class in a regular expression does not match anything and is almost certainly a mistake. Note that `[^]` (a negated empty class) is allowed since it matches any character.

Examples of **incorrect** code for this rule:

```javascript
var foo = /^abc[]/;
var foo = /foo[]bar/;
var foo = /[]]/;
```

Examples of **correct** code for this rule:

```javascript
var foo = /^abc[a-zA-Z]/;
var foo = /[^]/;
var foo = /[\\[]/;
var foo = /\\[]/;
```

## Original Documentation

- [ESLint no-empty-character-class](https://eslint.org/docs/latest/rules/no-empty-character-class)
