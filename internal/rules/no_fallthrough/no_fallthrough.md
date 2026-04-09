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

## Options

### `commentPattern`

A custom regular expression pattern to match fallthrough comments. By default matches `/falls?\s?through/i`.

```json
{ "no-fallthrough": ["error", { "commentPattern": "break[\\s\\w]*omitted" }] }
```

### `allowEmptyCase`

When set to `true`, allows case clauses containing only empty statements (`;`) to fall through without a comment.

```json
{ "no-fallthrough": ["error", { "allowEmptyCase": true }] }
```

### `reportUnusedFallthroughComment`

When set to `true`, reports fallthrough comments on cases that cannot actually fall through (e.g., cases ending with `break`).

```json
{ "no-fallthrough": ["error", { "reportUnusedFallthroughComment": true }] }
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-fallthrough
