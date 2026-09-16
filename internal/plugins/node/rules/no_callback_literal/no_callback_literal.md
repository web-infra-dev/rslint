# no-callback-literal

## Rule Details

Enforce the Node.js convention of passing an error as the first callback argument.
Calls to identifiers named `cb` or `callback` report literal values other than
`null`, including strings, numbers, booleans, arrays, objects, and template literals.

Examples of **incorrect** code for this rule:

```javascript
cb('Request failed');
cb({ message: 'Request failed' });
callback(0);
```

Examples of **correct** code for this rule:

```javascript
cb(undefined);
cb(null, result);
callback(new Error('Request failed'));
callback(error);
```

The rule checks syntax and allows expressions whose value may be an error. It does
not verify variable types. Calls without arguments and member calls such as
`object.callback('message')` are allowed.

This rule has no options.

## Original Documentation

- [eslint-plugin-n: no-callback-literal](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-callback-literal.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-callback-literal.js)
