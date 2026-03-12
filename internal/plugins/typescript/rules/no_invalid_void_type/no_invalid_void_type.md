# no-invalid-void-type

## Rule Details

Disallows `void` type outside of generic or return types. The `void` type in TypeScript means a function returns nothing, and is only meaningful as a return type of functions, methods, callable signatures, construct signatures, or as a generic type argument (e.g., `Promise<void>`). Using `void` in other positions such as variable types, parameter types, or union types is typically a mistake and can lead to confusing behavior.

Examples of **incorrect** code for this rule:

```typescript
let value: void;

function foo(arg: void) {}

type Union = string | void;

type KeyofVoid = keyof void;

let arr: void[];

let value = undefined as void;
```

Examples of **correct** code for this rule:

```typescript
function foo(): void {}

type Callback = () => void;

async function bar(): Promise<void> {}

type Result = void | never;

// Callable and construct signatures
interface Callable {
  (...args: string[]): void;
}

interface Constructable {
  new (...args: string[]): void;
}

// Function overloads - void in implementation return type is valid
function f(): void;
function f(x: string): string;
function f(x?: string): string | void {
  if (x !== undefined) {
    return x;
  }
}
```

## Options

### `allowInGenericTypeArguments`

- Type: `boolean | string[]`
- Default: `true`

When `true` (default), allows `void` as a type argument in any generic type (e.g., `Promise<void>`, `Map<string, void>`).

When set to an array of strings, only allows `void` as a type argument in the listed generic types. Supports dotted names (e.g., `['Promise', 'Ex.Mx.Tx']`).

When `false`, `void` is only valid as a direct return type.

### `allowAsThisParameter`

- Type: `boolean`
- Default: `false`

When `true`, allows `void` as the type of a `this` parameter in functions and methods.

```typescript
// Valid when allowAsThisParameter is true
function f(this: void) {}
class Test {
  method(this: void) {}
}
```

## Original Documentation

- [typescript-eslint no-invalid-void-type](https://typescript-eslint.io/rules/no-invalid-void-type)
