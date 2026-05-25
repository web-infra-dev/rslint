# no-extra-label

## Rule Details

This rule disallows labels that are only used on loops or switch statements
that have no nested breakable statement — in those cases a bare `break` /
`continue` already refers to the directly-enclosing loop or switch, so the
label adds no information and can confuse readers who expect labels to
control deeper nesting.

Examples of **incorrect** code for this rule:

```javascript
A: while (a) {
    break A;
}

B: for (let i = 0; i < 10; ++i) {
    break B;
}

C: switch (a) {
    case 0:
        break C;
}
```

Examples of **correct** code for this rule:

```javascript
while (a) {
    break;
}

A: {
    break A;
}

B: while (a) {
    while (b) {
        break B;
    }
}

C: switch (a) {
    case 0:
        while (b) {
            break C;
        }
}
```

## Options

This rule has no options.

## Original Documentation

- [ESLint rule: no-extra-label](https://eslint.org/docs/latest/rules/no-extra-label)
