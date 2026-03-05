# prefer-as-const

## Rule Details

Enforce the use of `as const` over literal type assertions. Using `as const` is preferred because it more clearly expresses intent, consistently applies to all literal values, and enables deeper immutability for objects and arrays.

This rule flags type assertions (`as` or angle-bracket style) and type annotations on variable or property declarations where the asserted/annotated type is a literal type that matches the literal value.

Examples of **incorrect** code for this rule:

```typescript
let foo = 'bar' as 'bar';
let baz = 1 as 1;
let qux = <'bar'>'bar';
let x: 10 = 10;
```

Examples of **correct** code for this rule:

```typescript
let foo = 'bar' as const;
let baz = 1 as const;
let arr = [1, 2, 3] as const;
```

## Original Documentation

- [typescript-eslint prefer-as-const](https://typescript-eslint.io/rules/prefer-as-const)
