# no-obj-calls

## Rule Details

Disallows calling global objects (`Math`, `JSON`, `Reflect`, `Atomics`, `Intl`) as functions or constructors. These are namespace objects that provide properties and methods but are not themselves callable. Attempting to call them will throw a `TypeError` at runtime.

Examples of **incorrect** code for this rule:

```javascript
var x = Math();
var y = JSON();
var z = Reflect();
var a = new Math();
var b = new JSON();
```

Examples of **correct** code for this rule:

```javascript
var x = Math.random();
var y = JSON.parse('{}');
var z = Reflect.get(obj, 'key');
var a = new Intl.Segmenter();
var b = Math.PI;
```

## Original Documentation

- [ESLint no-obj-calls](https://eslint.org/docs/latest/rules/no-obj-calls)
