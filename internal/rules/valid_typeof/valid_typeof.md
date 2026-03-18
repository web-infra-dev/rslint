<!-- cspell:ignore strnig undefned nunber fucntion -->

# valid-typeof

## Rule Details

Enforces comparing `typeof` expressions against valid string literals. The `typeof` operator can only return one of the following strings: `"undefined"`, `"object"`, `"boolean"`, `"number"`, `"string"`, `"function"`, `"symbol"`, `"bigint"`. Comparing a `typeof` expression against any other value is almost certainly a bug.

Examples of **incorrect** code for this rule:

```javascript
typeof foo === 'strnig';
typeof foo == 'undefned';
typeof bar != 'nunber';
typeof bar !== 'fucntion';
typeof foo === undefined;
```

Examples of **correct** code for this rule:

```javascript
typeof foo === 'string';
typeof bar == 'undefined';
typeof baz === 'object';
typeof qux !== 'function';
typeof foo === typeof bar;
```

## Options

### `requireStringLiterals`

When set to `true`, requires that `typeof` expressions are only compared to string literals or other `typeof` expressions, and disallows comparisons to any other value.

Examples of additional **incorrect** code with `{ "requireStringLiterals": true }`:

```javascript
typeof foo === undefined;
typeof foo === Object;
typeof foo === someVariable;
```

## Original Documentation

- [ESLint valid-typeof](https://eslint.org/docs/latest/rules/valid-typeof)
