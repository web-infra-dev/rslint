# no-invalid-void-type

## Rule Details

Disallows `void` type outside of generic or return types. The `void` type in TypeScript means a function returns nothing, and is only meaningful as a return type or as a generic type argument (e.g., `Promise<void>`). Using `void` in other positions such as variable types, parameter types, or union types is typically a mistake and can lead to confusing behavior.

Examples of **incorrect** code for this rule:

```typescript
let value: void;

function foo(arg: void) {}

type Union = string | void;
```

Examples of **correct** code for this rule:

```typescript
function foo(): void {}

type Callback = () => void;

async function bar(): Promise<void> {}

type Result = void | never;
```

## Original Documentation

- [typescript-eslint no-invalid-void-type](https://typescript-eslint.io/rules/no-invalid-void-type)
