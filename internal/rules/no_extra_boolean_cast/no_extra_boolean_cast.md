# no-extra-boolean-cast

## Rule Details

Disallows unnecessary boolean casts. Using `!!` (double negation) or `Boolean()` to convert a value to boolean is redundant when the value is already in a boolean context, such as the test of an `if` statement.

`new Boolean(x)` is never flagged because it produces a Boolean **object** (always truthy) rather than a primitive, so replacing it with a plain value would change semantics.

Examples of **incorrect** code for this rule:

```javascript
if (!!foo) {
}
while (!!foo) {}
do {} while (!!foo);
for (; !!foo; ) {}
!!foo ? bar : baz;
!!!foo;
if (Boolean(foo)) {
}
!Boolean(foo);
```

Examples of **correct** code for this rule:

```javascript
if (foo) {
}
while (foo) {}
var bar = !!foo;
var bar = Boolean(foo);
function baz() {
  return !!foo;
}
if (new Boolean(foo)) {
} // always truthy — not equivalent to `if (foo)`
```

## Options

This rule accepts a single options object.

### `enforceForLogicalOperands` (legacy)

When `true`, the rule also reports redundant boolean casts that are operands of `||` or `&&` when the overall logical expression is used in a boolean context.

```json
{ "no-extra-boolean-cast": ["error", { "enforceForLogicalOperands": true }] }
```

```javascript
if (x || !!y) {
} // reported
```

### `enforceForInnerExpressions`

A superset of `enforceForLogicalOperands`. Additionally reports redundant casts on the right-hand side of `??`, on the branches of ternaries, and on the last expression of a sequence (`a, b, c`).

```json
{ "no-extra-boolean-cast": ["error", { "enforceForInnerExpressions": true }] }
```

```javascript
if (x ?? !!y) {
} // reported
if (cond ? Boolean(a) : b) {
} // reported
if ((a, b, Boolean(c))) {
} // reported
```

## Original Documentation

- [ESLint no-extra-boolean-cast](https://eslint.org/docs/latest/rules/no-extra-boolean-cast)
