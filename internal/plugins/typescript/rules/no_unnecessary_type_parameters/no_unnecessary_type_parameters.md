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

## Original Documentation

[typescript-eslint: no-unnecessary-type-parameters](https://typescript-eslint.io/rules/no-unnecessary-type-parameters)
