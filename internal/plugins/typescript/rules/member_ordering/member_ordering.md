# member-ordering

## Rule Details

Require a consistent member declaration order.

A consistent ordering of fields, methods and constructors can make interfaces, type literals, classes and class expressions easier to read, navigate and edit.

This rule accepts an order configuration for each of the following AST node types:

- `default` — default ordering for all node types
- `classes` — ordering for class declarations
- `classExpressions` — ordering for class expressions
- `interfaces` — ordering for interface declarations
- `typeLiterals` — ordering for type literal declarations

Examples of **incorrect** code for this rule with the default configuration:

```typescript
interface Foo {
  B(): void;
  new (): Foo;
  A: string;
  [Z: string]: any;
}
```

Examples of **correct** code for this rule with the default configuration:

```typescript
interface Foo {
  [Z: string]: any;
  A: string;
  new (): Foo;
  B(): void;
}
```

## Original Documentation

https://typescript-eslint.io/rules/member-ordering
