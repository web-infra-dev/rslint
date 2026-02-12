# default-case

## Rule Details

Require `default` cases in `switch` statements. The rule also allows an opt-out comment such as `// no default`.

Examples of **incorrect** code for this rule:

```javascript
switch (a) {
  case 1:
    break;
}
```

Examples of **correct** code for this rule:

```javascript
switch (a) {
  case 1:
    break;
  default:
    break;
}

switch (a) {
  case 1:
    break;
  // no default
}
```

## Options

- `commentPattern`: A regular expression pattern for the opt-out comment. Default: `^no default$` (case-insensitive).

## Original Documentation

https://eslint.org/docs/latest/rules/default-case
