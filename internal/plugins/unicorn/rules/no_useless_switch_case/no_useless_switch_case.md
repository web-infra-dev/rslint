# no-useless-switch-case

## Rule Details

Disallows empty `case` clauses immediately before the final `default` clause in
a `switch` statement.

An empty case before the last `default` case is useless because execution falls
through to the same default body. Empty blocks and empty statements inside that
case still count as empty.

Examples of **incorrect** code for this rule:

```javascript
switch (foo) {
  case 1:
  default:
    handleDefaultCase();
    break;
}
```

```javascript
switch (foo) {
  case 1: {
  }
  default:
    handleDefaultCase();
    break;
}
```

Examples of **correct** code for this rule:

```javascript
switch (foo) {
  default:
    handleDefaultCase();
    break;
}
```

```javascript
switch (foo) {
  case 1:
  case 2:
    handleCase1And2();
    break;
}
```

```javascript
switch (foo) {
  case 1:
    handleCase1();
  default:
    handleDefaultCase();
    break;
}
```

## Original Documentation

- [eslint-plugin-unicorn: no-useless-switch-case](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/no-useless-switch-case.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/no-useless-switch-case.js)
