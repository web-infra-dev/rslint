# no-unused-labels

## Rule Details

This rule disallows labels that are declared but never used by a labeled
`break` or `continue` statement.

Examples of **incorrect** code for this rule:

```javascript
A: var foo = 0;

B: {
    foo();
}

C: for (let i = 0; i < 10; ++i) {
    foo();
}
```

Examples of **correct** code for this rule:

```javascript
A: {
    if (foo()) {
        break A;
    }
    bar();
}

B: for (let i = 0; i < 10; ++i) {
    if (foo()) {
        continue B;
    }
    bar();
}
```

## Options

This rule has no options.

## Original Documentation

- [ESLint rule: no-unused-labels](https://eslint.org/docs/latest/rules/no-unused-labels)
