# no-object-constructor

## Rule Details

Use of the `Object` constructor to construct a new empty object is generally discouraged in favor of object literal notation because of conciseness and because the `Object` global may be redefined. The exception is when the `Object` constructor is used to intentionally wrap a specified value which is passed as an argument.

This rule disallows calling the `Object` constructor without an argument.

Examples of **incorrect** code for this rule:

```javascript
Object();

new Object();
```

Examples of **correct** code for this rule:

```javascript
Object("foo");

const obj = { a: 1, b: 2 };

const isObject = (value) => value === Object(value);

const createObject = (Object) => new Object();
```

## Original Documentation

- [ESLint: no-object-constructor](https://eslint.org/docs/latest/rules/no-object-constructor)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-object-constructor.js)
