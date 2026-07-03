# prefer-arrow-callback

## Rule Details

Requires using arrow functions for callbacks when the function expression can be
replaced without changing `this`, `super`, `arguments`, or `new.target`
semantics.

Examples of **incorrect** code for this rule:

```javascript
foo(function (value) {
  return value;
});

foo(function () {
  return this.value;
}.bind(this));
```

Examples of **correct** code for this rule:

```javascript
foo((value) => {
  return value;
});

foo(() => {
  return this.value;
});

foo(function () {
  return this.value;
});
```

## Options

This rule accepts an options object with the following properties:

- `allowNamedFunctions` defaults to `false`. When `true`, named function
  expressions are allowed.
- `allowUnboundThis` defaults to `true`. When `false`, callbacks that reference
  their own `this` are still reported, but they are not automatically fixed.

Examples of **correct** code for this rule with `{ "allowNamedFunctions": true }`:

```json
{ "prefer-arrow-callback": ["error", { "allowNamedFunctions": true }] }
```

```javascript
foo(function namedCallback() {});
```

Examples of **incorrect** code for this rule with `{ "allowUnboundThis": false }`:

```json
{ "prefer-arrow-callback": ["error", { "allowUnboundThis": false }] }
```

```javascript
foo(function () {
  return this.value;
});
```

## Original Documentation

- [ESLint: prefer-arrow-callback](https://eslint.org/docs/latest/rules/prefer-arrow-callback)
