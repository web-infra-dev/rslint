# no-dupe-args

## Rule Details

Disallow duplicate arguments in `function` definitions. If more than one parameter has the same name in a function definition, the last occurrence "shadows" the preceding ones.

Examples of **incorrect** code for this rule:

```javascript
function foo(a, b, a) {
  console.log(a);
}

var bar = function (a, b, a) {};
```

Examples of **correct** code for this rule:

```javascript
function foo(a, b, c) {
  console.log(a, b, c);
}

var bar = function (a, b, c) {};
```

## Differences from ESLint

When a parameter name appears more than twice (e.g., `function foo(a, a, a)`), rslint reports an error on **each** duplicate occurrence (2 errors), while ESLint reports only once per duplicated name (1 error). This provides more precise diagnostic locations.

## Original Documentation

https://eslint.org/docs/latest/rules/no-dupe-args
