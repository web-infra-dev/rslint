# no-multi-str

## Rule Details

Disallows multiline strings created using a trailing backslash before a line break. This syntax was historically an undocumented feature of JavaScript and should be avoided.

Examples of **incorrect** code for this rule:

```javascript
var x =
  'Line 1 \
         Line 2';
```

Examples of **correct** code for this rule:

```javascript
var x = 'Line 1 ' + 'Line 2';

var x = `Line 1
         Line 2`;
```

## Original Documentation

- [ESLint no-multi-str](https://eslint.org/docs/latest/rules/no-multi-str)
