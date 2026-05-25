# no-proto

## Rule Details

Disallow the use of the `__proto__` property.

When an object is created with the `new` operator, `__proto__` is set to the original "prototype" property of the object's constructor function. `Object.getPrototypeOf` is the preferred method of getting the object's prototype. To change an object's prototype, use `Object.setPrototypeOf`.

Examples of **incorrect** code for this rule:

```javascript
var a = obj.__proto__;

var a = obj['__proto__'];

obj.__proto__ = b;

obj['__proto__'] = b;
```

Examples of **correct** code for this rule:

```javascript
var a = Object.getPrototypeOf(obj);

Object.setPrototypeOf(obj, b);

var c = { __proto__: a };
```

## Original Documentation

- [ESLint no-proto](https://eslint.org/docs/latest/rules/no-proto)
