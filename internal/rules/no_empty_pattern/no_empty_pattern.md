# no-empty-pattern

## Rule Details

Disallow empty destructuring patterns. Empty destructuring patterns do not create any variables and may be a sign of a mistake.

Examples of **incorrect** code for this rule:

```javascript
var {} = foo;
var [] = foo;
var {
  a: {},
} = foo;
var {
  a: [],
} = foo;
function foo({}) {}
function foo([]) {}
```

Examples of **correct** code for this rule:

```javascript
var { a } = foo;
var [a] = foo;
var { a = {} } = foo;
var {
  a: { b },
} = foo;
function foo({ a }) {}
function foo([a]) {}
```

## Options

- `allowObjectPatternsAsParameters`: If `true`, allows empty object patterns as function parameters. Default: `false`.

## Original Documentation

- [ESLint: no-empty-pattern](https://eslint.org/docs/latest/rules/no-empty-pattern)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-empty-pattern.js)

https://eslint.org/docs/latest/rules/no-empty-pattern
