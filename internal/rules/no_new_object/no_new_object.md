# no-new-object

## Rule Details

Disallow `Object` constructors. The object literal notation `{}` is preferable.

Examples of **incorrect** code for this rule:

```javascript
var myObject = new Object();

new Object();
```

Examples of **correct** code for this rule:

```javascript
var myObject = {};

var myObject = new CustomObject();

var foo = new foo.Object();
```

## Original Documentation

- [ESLint: no-new-object](https://eslint.org/docs/latest/rules/no-new-object)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-new-object.js)

https://eslint.org/docs/latest/rules/no-new-object
