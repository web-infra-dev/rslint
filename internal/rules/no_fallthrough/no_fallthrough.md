# no-fallthrough

## Rule Details

Disallow fallthrough of `case` statements. A case clause that has statements but does not end with a control flow statement (`break`, `return`, `throw`, `continue`) will fall through to the next case, which is usually a programming error. Empty case clauses (with no statements) are allowed by default.

A comment containing "falls through" or "fall through" (case-insensitive) between the current case and the next will suppress the warning.

Examples of **incorrect** code for this rule:

```javascript
switch (foo) {
  case 0:
    a();
  case 1:
    b();
    break;
}

switch (foo) {
  case 0:
    a();
  default:
    b();
}
```

Examples of **correct** code for this rule:

```javascript
switch (foo) {
  case 0:
    a();
    break;
  case 1:
    b();
    break;
}

switch (foo) {
  case 0:
  case 1:
    a();
    break;
}

switch (foo) {
  case 0:
    a();
  /* falls through */
  case 1:
    b();
    break;
}

function bar() {
  switch (foo) {
    case 0:
      a();
      return;
    case 1:
      b();
  }
}
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-fallthrough
