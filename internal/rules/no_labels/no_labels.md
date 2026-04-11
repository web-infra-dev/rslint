# no-labels

## Rule Details

Disallow labeled statements. Labels tend to be used only rarely and are frowned upon as a remedial form of flow control that is more error prone and harder to understand.

This rule aims to eliminate the use of labeled statements in JavaScript and reports whenever a labeled statement is encountered and whenever `break` or `continue` are used with a label.

Examples of **incorrect** code for this rule:

```javascript
label: while (true) {}

label: while (true) {
  break label;
}

label: while (true) {
  continue label;
}
```

Examples of **correct** code for this rule:

```javascript
var f = { label: foo() };

while (true) {}

while (true) {
  break;
}

while (true) {
  continue;
}
```

## Options

- `allowLoop` (boolean, default `false`): When `true`, allows labels attached to loop statements.
- `allowSwitch` (boolean, default `false`): When `true`, allows labels attached to switch statements.

Examples of **correct** code with `{ "allowLoop": true }`:

```javascript
A: while (a) {
  break A;
}

A: do {
  if (b) {
    break A;
  }
} while (a);

A: for (var a in obj) {
  for (;;) {
    switch (a) {
      case 0:
        continue A;
    }
  }
}
```

Examples of **correct** code with `{ "allowSwitch": true }`:

```javascript
A: switch (a) {
  case 0:
    break A;
}
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-labels
