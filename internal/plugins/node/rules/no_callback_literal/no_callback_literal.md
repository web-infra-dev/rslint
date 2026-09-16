# no-callback-literal

## Rule Details

Enforce the Node.js convention of passing an error as the first callback argument.
Calls to identifiers named `cb` or `callback` report literal values other than
`null`, including strings, numbers, booleans, regular expressions, arrays, objects,
and template literals.

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

## Differences from upstream

rslint reports regular expression literals in the first callback argument,
including newer syntax such as `cb(/(?i:error)/)`. When ESLint runs on a Node.js
version that does not support that syntax, such as Node.js 22.18.0, the upstream
rule may allow the same code. Use an `Error` for failures or put the regular
expression in the result position, for example `cb(null, /(?i:error)/)`.

## Original Documentation

- [eslint-plugin-n: no-callback-literal](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-callback-literal.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-callback-literal.js)
