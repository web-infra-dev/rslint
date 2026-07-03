# prefer-regex-literals

## Rule Details

This rule disallows `RegExp` constructor calls when the same regular expression
can be written as a literal.

Examples of **incorrect** code for this rule:

```javascript
new RegExp("abc");
RegExp("abc", "u");
new RegExp(String.raw`^\d\.$`);
```

Examples of **correct** code for this rule:

```javascript
/abc/;
/abc/u;

new RegExp(pattern);
RegExp("abc", flags);
new RegExp(prefix + "abc");
```

## Options

This rule accepts an options object with the following property:

- `disallowRedundantWrapping` (`boolean`, default `false`) — when `true`,
  additionally reports regex literals that are unnecessarily wrapped in a
  `RegExp` constructor.

To enable this option:

```json
{ "prefer-regex-literals": ["error", { "disallowRedundantWrapping": true }] }
```

Examples of **incorrect** code for this rule with
`{ "disallowRedundantWrapping": true }`:

```javascript
new RegExp(/abc/);
new RegExp(/abc/, "u");
```

Examples of **correct** code for this rule with
`{ "disallowRedundantWrapping": true }`:

```javascript
/abc/;
/abc/u;
new RegExp(/abc/, flags);
```

## Differences from ESLint

- rslint does not vary suggestions by configured ECMAScript version; a
  diagnostic is still reported when the pattern arguments are static.

## Original Documentation

- [ESLint prefer-regex-literals](https://eslint.org/docs/latest/rules/prefer-regex-literals)
