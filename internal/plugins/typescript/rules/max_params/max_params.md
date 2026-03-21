# max-params

## Rule Details

Enforce a maximum number of parameters in function-like declarations.

Examples of **incorrect** code for this rule:

```typescript
function foo(a, b, c, d) {}

class Foo {
  method(this: Foo, a, b, c) {}
}
```

Examples of **correct** code for this rule:

```typescript
function foo(a, b, c) {}

class Foo {
  method(this: void, a, b, c) {}
}
```

## Options

- `max` (`number`, default: `3`): Maximum number of parameters.
- `maximum` (`number`, deprecated): Alias of `max`.
- `countVoidThis` (`boolean`, default: `false`): When `false`, a leading `this: void` parameter is not counted.

The numeric shorthand is also supported:

```typescript
// equivalent to { max: 4 }
['@typescript-eslint/max-params', 2, 4];
```

## Original Documentation

https://typescript-eslint.io/rules/max-params
