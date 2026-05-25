# no-octal-escape

## Rule Details

Disallows octal escape sequences in string literals.

As of the ECMAScript 5 specification, octal escape sequences in string literals are deprecated and should not be used. Unicode escape sequences should be used instead.

Examples of **incorrect** code for this rule:

```javascript
var foo = 'Copyright \251';
var foo = '\1';
var foo = '\01';
var foo = '\08';
```

Examples of **correct** code for this rule:

```javascript
var foo = 'Copyright \u00A9';
var foo = '\x51';
var foo = '\0';
var foo = '\\1';
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-octal-escape
