# no-else-return

## Rule Details

Disallow `else` blocks after `return` statements in `if` statements. When every
preceding branch returns, the `else` block is unnecessary and its contents can
be placed after the `if` statement.

Examples of **incorrect** code for this rule:

```javascript
function foo() {
  if (x) {
    return y;
  } else {
    return z;
  }
}
```

```javascript
function foo() {
  if (x) {
    return y;
  } else if (z) {
    return w;
  } else {
    return q;
  }
}
```

Examples of **correct** code for this rule:

```javascript
function foo() {
  if (x) {
    return y;
  }

  return z;
}
```

```javascript
function foo() {
  if (x) {
    doSomething();
  } else {
    return y;
  }
}
```

## Options

This rule has an object option:

- `allowElseIf` (default: `true`): allows `else if` blocks after a `return`.

Examples of **correct** code for this rule with `{ "allowElseIf": true }`:

```json
{ "no-else-return": ["error", { "allowElseIf": true }] }
```

```javascript
function foo() {
  if (error) {
    return "failed";
  } else if (loading) {
    return "loading";
  }
}
```

Examples of **incorrect** code for this rule with `{ "allowElseIf": false }`:

```json
{ "no-else-return": ["error", { "allowElseIf": false }] }
```

```javascript
function foo() {
  if (error) {
    return "failed";
  } else if (loading) {
    return "loading";
  }
}
```

## Original Documentation

- [ESLint: no-else-return](https://eslint.org/docs/latest/rules/no-else-return)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-else-return.js)

https://eslint.org/docs/latest/rules/no-else-return
