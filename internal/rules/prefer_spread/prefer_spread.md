# prefer-spread

## Rule Details

Suggest using the spread operator instead of `.apply()`.

Before ES2015, `Function.prototype.apply()` was the only way to call a variadic
function with an array of arguments. With the spread operator (`...`),
`function(...args)` achieves the same effect more concisely and also works in
`new` expressions, which `.apply()` does not support.

This rule flags `.apply()` calls that are interchangeable with a spread call:

- The second argument of `.apply()` is neither an array literal nor a spread
  element (those forms already behave like a spread call).
- The first argument preserves the `this` binding of the applied function:
  - When the function is not a member expression, only `null` / `undefined` /
    `void 0` pass (otherwise the `this` binding may change on migration).
  - When the function is a member expression (e.g. `obj.foo.apply(obj, args)`),
    the first argument must produce the same token stream as the member's
    object (e.g. `obj` on both sides).

Examples of **incorrect** code for this rule:

```javascript
foo.apply(undefined, args);
foo.apply(null, args);
obj.foo.apply(obj, args);
```

Examples of **correct** code for this rule:

```javascript
// The `this` binding is changed deliberately
foo.apply(obj, args);
obj.foo.apply(null, args);
obj.foo.apply(otherObj, args);

// The second argument is not variadic
foo.apply(undefined, [1, 2, 3]);
obj.foo.apply(obj, [1, 2, 3]);

// Already using a spread call
obj.foo(...args);
```

## Original Documentation

- [ESLint prefer-spread](https://eslint.org/docs/latest/rules/prefer-spread)
