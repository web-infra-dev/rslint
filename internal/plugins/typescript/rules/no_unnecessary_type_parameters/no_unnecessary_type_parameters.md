# no-unnecessary-type-parameters

## Rule Details

Disallows type parameters that aren't used multiple times in a function, method, or class signature. A type parameter relates two or more positions in a signature; if it only appears once, it isn't relating anything, and the concrete type it stands for can be written directly instead.

Examples of **incorrect** code for this rule:

```typescript
function second<A, B>(a: A, b: B): B {
  return b;
}

function parseJSON<T>(input: string): T {
  return JSON.parse(input);
}

function printProperty<T, K extends keyof T>(obj: T, key: K) {
  console.log(obj[key]);
}
```

Examples of **correct** code for this rule:

```typescript
function second<B>(a: unknown, b: B): B {
  return b;
}

function parseJSON(input: string): unknown {
  return JSON.parse(input);
}

// T appears twice: once as the parameter type, once as the return type.
function identity<T>(arg: T): T {
  return arg;
}

// T appears twice: `keyof T` and the inferred return type (`T[K]`).
// K appears twice: `key: K` and the inferred return type (`T[K]`).
function getProperty<T, K extends keyof T>(obj: T, key: K) {
  return obj[key];
}
```

## Differences from ESLint

rslint resolves a few signature positions that the TypeScript public API keeps out of reach of upstream `@typescript-eslint/no-unnecessary-type-parameters`, and counts a type parameter appearing there as used. That makes for slightly fewer reports, all of them ones upstream raises in error.

The `replaceUsagesWithConstraint` suggestion parenthesizes the constraint wherever the type grammar calls for it, so applying it always leaves you with the same type the code had before.

## Original Documentation

- [typescript-eslint: no-unnecessary-type-parameters](https://typescript-eslint.io/rules/no-unnecessary-type-parameters)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/no-unnecessary-type-parameters.ts)
