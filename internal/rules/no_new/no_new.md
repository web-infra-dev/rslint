# no-new

## Rule Details

Disallows the use of `new` operators outside of assignments or comparisons. The goal of `new` with a constructor is to create a new object of a particular type and assign or compare that object. A `new` expression used as a standalone statement discards the resulting object, which usually means the constructor should have been a plain function call instead.

Examples of **incorrect** code for this rule:

```javascript
new Thing();
```

Examples of **correct** code for this rule:

```javascript
var thing = new Thing();

Thing();
```

## Original Documentation

- [ESLint no-new](https://eslint.org/docs/latest/rules/no-new)
