# no-return-assign

## Rule Details

This rule aims to eliminate assignments from `return` statements, because it is difficult to tell whether the author intended an assignment or a mistyped comparison.

Examples of **incorrect** code for this rule:

```javascript
function doSomething() {
  return foo = bar + 2;
}

function doSomethingElse() {
  return foo += 2;
}

const foo = (a, b) => a = b;

const bar = (a, b, c) => (a = b, c == b);

function doSomethingMore() {
  return foo = bar && foo > 0;
}
```

Examples of **correct** code for this rule:

```javascript
function doSomething() {
  return foo == bar + 2;
}

function doSomethingMore() {
  return (foo = bar + 2);
}

const foo = (a, b) => (a = b);

const bar = (a, b, c) => ((a = b), c == b);

function doAnotherThing() {
  return (foo = bar) && foo > 0;
}
```

## Options

This rule takes a single string option:

- `"except-parens"` (default) — disallow assignments in `return` statements unless they are enclosed in parentheses.
- `"always"` — disallow all assignments in `return` statements, even when parenthesised.

Examples of **incorrect** code for this rule with `"always"`:

```json
{ "no-return-assign": ["error", "always"] }
```

```javascript
function doSomething() {
  return foo = bar + 2;
}

function doSomethingMore() {
  return (foo = bar + 2);
}
```

Examples of **correct** code for this rule with `"always"`:

```json
{ "no-return-assign": ["error", "always"] }
```

```javascript
function doSomething() {
  return foo == bar + 2;
}
```

## Original Documentation

- [ESLint rule: no-return-assign](https://eslint.org/docs/latest/rules/no-return-assign)
