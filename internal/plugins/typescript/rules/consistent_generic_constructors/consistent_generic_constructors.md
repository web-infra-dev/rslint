# consistent-generic-constructors

## Rule Details

Enforce specifying generic type arguments on the type annotation or the constructor of a variable declaration. When constructing a generic class, the type arguments can be placed either on the left-hand side (type annotation) or the right-hand side (constructor call). This rule enforces consistency.

The rule supports two modes: `"constructor"` (default) prefers type arguments on the constructor call, and `"type-annotation"` prefers them on the type annotation.

Examples of **incorrect** code for this rule (with default `"constructor"` option):

```typescript
const map: Map<string, number> = new Map();
const set: Set<string> = new Set();
```

Examples of **correct** code for this rule (with default `"constructor"` option):

```typescript
const map = new Map<string, number>();
const set = new Set<string>();
const map2: Map<string, number> = new Map<string, number>();
```

## Original Documentation

- [typescript-eslint consistent-generic-constructors](https://typescript-eslint.io/rules/consistent-generic-constructors)
