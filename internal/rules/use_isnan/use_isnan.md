# use-isnan

## Rule Details

Requires calls to `isNaN()` when checking for `NaN`. Because `NaN` is unique in JavaScript in that it is not equal to anything, including itself, the results of comparisons to `NaN` are confusing: `NaN === NaN` is `false`. Therefore, use `Number.isNaN()` or the global `isNaN()` function to test whether a value is `NaN`.

## Options

- `enforceForSwitchCase` (default: `true`): Disallows `switch(NaN)` and `case NaN:` in switch statements.
- `enforceForIndexOf` (default: `false`): Disallows calling `indexOf` and `lastIndexOf` with `NaN` as an argument.

Examples of **incorrect** code for this rule:

```javascript
if (foo == NaN) {
}

if (foo === NaN) {
}

if (foo !== NaN) {
}

switch (NaN) {
  case foo:
    break;
}

switch (foo) {
  case NaN:
    break;
}
```

Examples of **correct** code for this rule:

```javascript
if (isNaN(foo)) {
}

if (Number.isNaN(foo)) {
}

if (!isNaN(foo)) {
}
```

## Original Documentation

- [ESLint use-isnan](https://eslint.org/docs/latest/rules/use-isnan)
