# method-signature-style

## Rule Details

Enforces using a particular method signature syntax in interfaces and type literals.

There are two styles for declaring method signatures in TypeScript:

- **Method shorthand**: `func(arg: string): number;`
- **Property**: `func: (arg: string) => number;`

The key difference: with `strictFunctionTypes` enabled, method parameters are checked less strictly, while function property parameters are checked more strictly. This makes function properties more type-safe.

### Options

- `"property"` (default): Enforces function property signature syntax.
- `"method"`: Enforces method shorthand signature syntax.

### `"property"` (default)

Examples of **incorrect** code:

```typescript
interface T1 {
  func(arg: string): number;
}
```

Examples of **correct** code:

```typescript
interface T1 {
  func: (arg: string) => number;
}
```

### `"method"`

Examples of **incorrect** code:

```typescript
interface T1 {
  func: (arg: string) => number;
}
```

Examples of **correct** code:

```typescript
interface T1 {
  func(arg: string): number;
}
```

## Original Documentation

https://typescript-eslint.io/rules/method-signature-style
