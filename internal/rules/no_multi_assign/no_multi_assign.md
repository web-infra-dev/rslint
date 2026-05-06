# no-multi-assign

## Rule Details

This rule disallows chained assignment expressions within a single statement, such as `a = b = c`. Chained assignments are often a sign of a typo (`foo = bar == 0` was meant) and they hide whether each name is being declared or merely reassigned, which can lead to surprising scope and `const`-vs-`let` mistakes.

Examples of **incorrect** code for this rule:

```javascript
var a = b = c = 5;

const foo = bar = "baz";

let a =
  b =
    c;

class Foo {
  a = b = 10;
}

a = b = "quux";
```

Examples of **correct** code for this rule:

```javascript
var a = 5;
var b = 5;
var c = 5;

const foo = "baz";
const bar = "baz";

let a = 5;
let b = 5;
let c = 5;

class Foo {
  a = 10;
  b = 10;
}

a = "quux";
b = "quux";
```

## Options

This rule has an object option:

- `"ignoreNonDeclaration"`: When set to `true`, allows chained assignments that do not introduce new declarations (i.e. plain `AssignmentExpression`s). Defaults to `false`.

Examples of **correct** code for this rule with `{ "ignoreNonDeclaration": true }`:

```json
{ "no-multi-assign": ["error", { "ignoreNonDeclaration": true }] }
```

```javascript
let a;
let b;
a = b = "baz";

const x = {};
const y = {};
x.one = y.one = 1;
```

Examples of **incorrect** code for this rule with `{ "ignoreNonDeclaration": true }`:

```json
{ "no-multi-assign": ["error", { "ignoreNonDeclaration": true }] }
```

```javascript
let a = b = "baz";

const foo = bar = 1;

class Foo {
  a = b = 10;
}
```

## Original Documentation

- [ESLint no-multi-assign](https://eslint.org/docs/latest/rules/no-multi-assign)
