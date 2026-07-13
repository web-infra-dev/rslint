# no-buffer-constructor

## Rule Details

This rule disallows the use of the `Buffer()` constructor.

In Node.js, the `Buffer` constructor is deprecated because it is a source of security and usability issues. Instead, use `Buffer.from()`, `Buffer.alloc()`, or `Buffer.allocUnsafe()`.

Examples of **incorrect** code for this rule:

```javascript
Buffer(5);
new Buffer(5);
Buffer([1, 2, 3]);
new Buffer([1, 2, 3]);
```

Examples of **correct** code for this rule:

```javascript
Buffer.alloc(5);
Buffer.allocUnsafe(5);
new Buffer.Foo();
Buffer.from([1, 2, 3]);
```

## Original Documentation

[ESLint documentation](https://eslint.org/docs/latest/rules/no-buffer-constructor)
