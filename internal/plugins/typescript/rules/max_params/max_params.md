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

## Original Documentation

https://typescript-eslint.io/rules/max-params
