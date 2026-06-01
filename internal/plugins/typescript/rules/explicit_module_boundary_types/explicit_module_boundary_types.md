# explicit-module-boundary-types

Require explicit return and argument types on exported functions' and classes' public class methods.

## Rule Details

Code that is part of a module's public surface — exported functions, exported class methods, default-exported expressions — is consumed by other modules whose authors can't see your implementation. Annotating the parameter and return types at that boundary documents the contract, lets editors offer accurate help, and prevents downstream callers from accidentally relying on the inferred type of an internal helper.

This rule reports any exported function, exported class method, or value reachable via an export reference that omits its parameter types or return type.

Examples of **incorrect** code for this rule:

```typescript
export function test(a: number, b: number) {
  return;
}

export var arrowFn = () => 'test';

export function fn(test): string {
  return '123';
}

export class Test {
  method() {
    return;
  }
}
```

Examples of **correct** code for this rule:

```typescript
export function test(a: number, b: number): void {
  return;
}

export var arrowFn = (): string => 'test';

export function fn(test: string): string {
  return '123';
}

export class Test {
  method(): void {
    return;
  }
}
```

## Options

This rule accepts an options object with the following properties:

- `allowArgumentsExplicitlyTypedAsAny` (default `false`): permit parameters explicitly annotated as `any`.
- `allowDirectConstAssertionInArrowFunctions` (default `true`): skip the return-type check on body-less arrow functions whose result is `as const` (optionally followed by `satisfies T`). Parameters still must be typed.
- `allowedNames` (default `[]`): list of function or method names to skip entirely (both return type and parameters).
- `allowHigherOrderFunctions` (default `true`): skip the return-type check on functions that immediately return another function expression, as long as the inner function has a return type.
- `allowOverloadFunctions` (default `false`): skip the return-type check on the implementation of an overloaded function/method.
- `allowTypedFunctionExpressions` (default `true`): skip the return-type check on function expressions whose surrounding context already supplies a type (variable annotation, type assertion, typed property, JSX attribute, function argument, …).

### `allowArgumentsExplicitlyTypedAsAny`

Examples of **incorrect** code with `{ "allowArgumentsExplicitlyTypedAsAny": false }` (the default):

```json
{ "@typescript-eslint/explicit-module-boundary-types": ["error", { "allowArgumentsExplicitlyTypedAsAny": false }] }
```

```typescript
export function foo(foo: any): void {}
```

Examples of **correct** code with `{ "allowArgumentsExplicitlyTypedAsAny": true }`:

```json
{ "@typescript-eslint/explicit-module-boundary-types": ["error", { "allowArgumentsExplicitlyTypedAsAny": true }] }
```

```typescript
export function foo(foo: any): void {}
```

### `allowDirectConstAssertionInArrowFunctions`

Examples of **correct** code with the default `{ "allowDirectConstAssertionInArrowFunctions": true }`:

```typescript
export const func1 = (value: number) => ({ type: 'X', value }) as const;
```

### `allowedNames`

Examples of **correct** code with `{ "allowedNames": ["func1"] }`:

```json
{ "@typescript-eslint/explicit-module-boundary-types": ["error", { "allowedNames": ["func1"] }] }
```

```typescript
export const func1 = (value: number) => value;
```

### `allowHigherOrderFunctions`

Examples of **correct** code with the default `{ "allowHigherOrderFunctions": true }`:

```typescript
export const fn = () => (n: number): string => String(n);
```

### `allowOverloadFunctions`

Examples of **correct** code with `{ "allowOverloadFunctions": true }`:

```json
{ "@typescript-eslint/explicit-module-boundary-types": ["error", { "allowOverloadFunctions": true }] }
```

```typescript
export function test(a: string): string;
export function test(a: number): number;
export function test(a: unknown) {
  return a;
}
```

### `allowTypedFunctionExpressions`

Examples of **correct** code with the default `{ "allowTypedFunctionExpressions": true }`:

```typescript
export var arrowFn: Foo = () => 'test';
const x = (() => {}) as Foo;
```

## Original Documentation

- [typescript-eslint explicit-module-boundary-types](https://typescript-eslint.io/rules/explicit-module-boundary-types)
