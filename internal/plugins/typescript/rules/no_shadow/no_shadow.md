---
description: 'Disallow variable declarations from shadowing variables declared in the outer scope.'
---

# `@typescript-eslint/no-shadow`

Extends the ESLint core [`no-shadow`](https://eslint.org/docs/latest/rules/no-shadow) rule with TypeScript-aware behavior.

## Rule Details

Shadowing occurs when a local variable shares the same name as a variable in its containing scope. The TypeScript variant additionally understands type-only declarations and the related `typeof` interplay between values and types.

Examples of **incorrect** code for this rule:

```ts
type Foo = number;
function b() {
  type Foo = string;
}
```

Examples of **correct** code for this rule:

```ts
type Foo = number;
function b() {
  type Bar = string;
}
```

## Options

This rule accepts the same options as the ESLint core `no-shadow` rule with these additions and default differences:

### `hoist`

In addition to the core values (`"functions"`, `"all"`, `"never"`), this rule supports:

- `"types"`: report shadowing of an outer type or interface that appears later.
- `"functions-and-types"`: report shadowing of outer function or type declarations that appear later.

**Default**: `"functions-and-types"` (the core ESLint default is `"functions"`).

### `ignoreTypeValueShadow`

When `true`, a value declaration and a type declaration that share a name are not considered shadowing (the two live in different namespaces and a `typeof` is required to bridge them).

**Default**: `true`.

```json
{ "@typescript-eslint/no-shadow": ["error", { "ignoreTypeValueShadow": false }] }
```

```ts
type Foo = number;
function f() {
  const Foo = 1;
}
```

### `ignoreFunctionTypeParameterNameValueShadow`

When `true`, parameters declared inside a function type (for example `(x: string) => void`) do not report when they shadow an outer value binding.

**Default**: `true`.

```json
{ "@typescript-eslint/no-shadow": ["error", { "ignoreFunctionTypeParameterNameValueShadow": false }] }
```

```ts
const test = 1;
type Func = (test: string) => typeof test;
```

### `allow`, `builtinGlobals`, `ignoreOnInitialization`

Same semantics and defaults as the ESLint core rule.

## Original Documentation

[https://typescript-eslint.io/rules/no-shadow](https://typescript-eslint.io/rules/no-shadow)
