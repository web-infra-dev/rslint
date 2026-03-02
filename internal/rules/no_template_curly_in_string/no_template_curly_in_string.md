# no-template-curly-in-string

## Rule Details

Disallows template literal placeholder syntax (`${expression}`) inside regular strings. This is almost always a mistake where the developer intended to use a template literal (backtick-delimited string) but accidentally used single or double quotes instead, so the placeholder is treated as a literal string rather than being interpolated.

Examples of **incorrect** code for this rule:

```javascript
var greeting = 'Hello, ${name}!';

var query = 'SELECT * FROM ${table}';

var msg = 'The value is ${a + b}';
```

Examples of **correct** code for this rule:

```javascript
var greeting = `Hello, ${name}!`;

var query = `SELECT * FROM ${table}`;

var literal = 'This is a dollar sign: ${}'; // intentional
```

## Original Documentation

- [ESLint no-template-curly-in-string](https://eslint.org/docs/latest/rules/no-template-curly-in-string)
