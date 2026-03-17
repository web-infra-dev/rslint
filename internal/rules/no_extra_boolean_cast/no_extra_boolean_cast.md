# no-extra-boolean-cast

## Rule Details

Disallows unnecessary boolean casts. Using `!!` (double negation) or `Boolean()` to convert a value to boolean is redundant when the value is already in a boolean context, such as the test of an `if` statement.

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
```

## Original Documentation

- [ESLint no-extra-boolean-cast](https://eslint.org/docs/latest/rules/no-extra-boolean-cast)
