# max-params

Enforce a maximum number of parameters in function definitions.

## Rule Details

Functions that take many parameters are usually harder to read and maintain than
functions that take fewer. This rule reports any function whose parameter list
exceeds the configured maximum (default 3).

The typescript-eslint variant adds the `countVoidThis` option for the TypeScript
`this: void` parameter, which is a type annotation rather than a real argument.
By default, `this: void` is excluded from the parameter count.

Examples of **incorrect** code for this rule:

```javascript
function foo(a, b, c, d) {}

const bar = (a, b, c, d) => {};

class Foo {
  method(this: Foo, a, b, c) {}
}
```

Examples of **correct** code for this rule:

```javascript
function foo(a, b, c) {}

const bar = (a, b, c) => {};

class Foo {
  method(this: void, a, b, c) {}
}
```

## Options

The rule accepts an options object:

```json
{ "@typescript-eslint/max-params": ["error", { "max": 3 }] }
```

- `max` (default `3`): the maximum number of parameters allowed.
- `maximum`: deprecated alias for `max`.
- `countVoidThis` (default `false`): if `true`, count a `this: void` parameter
  toward the limit.

Examples of **incorrect** code with `{ "max": 2 }`:

```json
{ "@typescript-eslint/max-params": ["error", { "max": 2 }] }
```

```javascript
function foo(a, b, c) {}
```

Examples of **correct** code with `{ "max": 2 }`:

```json
{ "@typescript-eslint/max-params": ["error", { "max": 2 }] }
```

```javascript
function foo(a, b) {}
```

Examples of **incorrect** code with `{ "countVoidThis": true, "max": 2 }`:

```json
{ "@typescript-eslint/max-params": ["error", { "countVoidThis": true, "max": 2 }] }
```

```javascript
class Foo {
  method(this: void, a, b) {}
}
```

## Original Documentation

- [`@typescript-eslint/max-params`](https://typescript-eslint.io/rules/max-params)
- [`max-params` (ESLint core)](https://eslint.org/docs/latest/rules/max-params)
