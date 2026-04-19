# symbol-description

## Rule Details

Requires a description when creating a `Symbol`. A description makes logged and debugged symbols easier to identify.

Examples of **incorrect** code for this rule:

```javascript
var foo = Symbol();
```

Examples of **correct** code for this rule:

```javascript
var foo = Symbol("some description");

var someString = "some description";
var bar = Symbol(someString);
```

## Original Documentation

- [ESLint symbol-description](https://eslint.org/docs/latest/rules/symbol-description)
