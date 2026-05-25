# no-self-compare

## Rule Details

Disallows comparing a value to itself using any of the equality or relational operators (`===`, `==`, `!==`, `!=`, `>`, `<`, `>=`, `<=`). Such a comparison is typically a typo (the programmer likely meant a different operand) and is either always true, always false, or always `NaN`-sensitive, so the check is pointless.

Examples of **incorrect** code for this rule:

```javascript
var x = 10;
if (x === x) {
}

if (x !== x) {
}

if (foo.bar().baz.qux >= foo.bar().baz.qux) {
}
```

Examples of **correct** code for this rule:

```javascript
var x = 10;
var y = 10;
if (x === y) {
}

if (foo.bar.baz === foo.bar.qux) {
}

class C {
  #field;
  foo() {
    // Property access via private identifier and bracket string literal
    // are structurally distinct and allowed.
    return this.#field === this["#field"];
  }
}
```

## Original Documentation

- [ESLint no-self-compare](https://eslint.org/docs/latest/rules/no-self-compare)
