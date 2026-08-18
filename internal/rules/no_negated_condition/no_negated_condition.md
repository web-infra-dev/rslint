# no-negated-condition

## Rule Details

Negated conditions are more difficult to understand. Code can be made more readable by inverting the condition instead.

This rule disallows negated conditions in either of the following:

- `if` statements which have an `else` branch
- ternary expressions

Examples of **incorrect** code for this rule:

```javascript
if (!a) {
  doSomething();
} else {
  doSomethingElse();
}

if (a != b) {
  doSomething();
} else {
  doSomethingElse();
}

if (a !== b) {
  doSomething();
} else {
  doSomethingElse();
}

!a ? c : b;
```

Examples of **correct** code for this rule:

```javascript
if (!a) {
  doSomething();
}

if (!a) {
  doSomething();
} else if (b) {
  doSomething();
}

if (a != b) {
  doSomething();
}

a ? b : c;
```

## Original Documentation

- [ESLint: no-negated-condition](https://eslint.org/docs/latest/rules/no-negated-condition)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-negated-condition.js)
