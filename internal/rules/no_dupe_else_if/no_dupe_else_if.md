# no-dupe-else-if

## Rule Details

Disallow duplicate conditions in if-else-if chains. If an `else if` condition is identical to a previous condition in the same chain, the branch can never execute.

Examples of **incorrect** code for this rule:

```javascript
if (a) {
  foo();
} else if (a) {
  bar();
}

if (a) {
  foo();
} else if (b) {
  bar();
} else if (a) {
  baz();
}
```

Examples of **correct** code for this rule:

```javascript
if (a) {
  foo();
} else if (b) {
  bar();
}

if (a === 1) {
  foo();
} else if (a === 2) {
  bar();
}

if (a) {
  foo();
} else if (b) {
  bar();
} else {
  baz();
}
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-dupe-else-if
