# no-nested-ternary

## Rule Details

Disallows nested ternary expressions. Nesting ternary expressions can make code more difficult to understand; prefer an `if` statement or extract the logic into named variables.

Examples of **incorrect** code for this rule:

```javascript
var thing = foo ? bar : baz === qux ? quxx : foobar;

foo ? (baz === qux ? quxx : foobar) : bar;
```

Examples of **correct** code for this rule:

```javascript
var thing = foo ? bar : foobar;

var thing;
if (foo) {
  thing = bar;
} else if (baz === qux) {
  thing = quxx;
} else {
  thing = foobar;
}
```

## Original Documentation

- [ESLint no-nested-ternary](https://eslint.org/docs/latest/rules/no-nested-ternary)
