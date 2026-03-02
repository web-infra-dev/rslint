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

## Original Documentation

https://eslint.org/docs/latest/rules/no-dupe-args
