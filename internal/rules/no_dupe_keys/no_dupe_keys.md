# no-dupe-keys

## Rule Details

Disallow duplicate keys in object literals. Multiple properties with the same key in object literals can cause unexpected behavior.

Examples of **incorrect** code for this rule:

```javascript
var foo = {
  bar: 'baz',
  bar: 'qux',
};

var foo = {
  bar: 'baz',
  bar: 'qux',
};
```

Examples of **correct** code for this rule:

```javascript
var foo = {
  bar: 'baz',
  qux: 'quux',
};

// getter and setter with same name is valid
var foo = {
  get bar() {},
  set bar(v) {},
};
```

## Original Documentation

- [ESLint: no-dupe-keys](https://eslint.org/docs/latest/rules/no-dupe-keys)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-dupe-keys.js)

https://eslint.org/docs/latest/rules/no-dupe-keys
