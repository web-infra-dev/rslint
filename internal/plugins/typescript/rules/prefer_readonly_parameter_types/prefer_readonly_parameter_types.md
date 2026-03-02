# prefer-readonly-parameter-types

## Rule Details

Require function parameters to be typed as `readonly` to prevent accidental mutation of inputs. Mutating function arguments can lead to confusing, hard to debug behavior. This rule flags function parameters whose types are not readonly, encouraging the use of `Readonly<T>`, `readonly` arrays, and other immutable type constructs.

Examples of **incorrect** code for this rule:

```typescript
function foo(arr: string[]) {}
function bar(obj: { prop: string }) {}
function baz(arg: Set<string>) {}
```

Examples of **correct** code for this rule:

```typescript
function foo(arr: readonly string[]) {}
function bar(obj: Readonly<{ prop: string }>) {}
function baz(arg: ReadonlySet<string>) {}
function qux(value: string) {} // primitives are always readonly
```

## Original Documentation

- [typescript-eslint prefer-readonly-parameter-types](https://typescript-eslint.io/rules/prefer-readonly-parameter-types)
