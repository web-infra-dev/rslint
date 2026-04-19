# no-unnecessary-type-constraint

## Rule Details

This rule disallows unnecessary constraints on generic types. Type parameters (`<T>`) default to `unknown`, so constraining a generic type parameter to `any` or `unknown` has no effect.

Examples of **incorrect** code for this rule:

```typescript
interface FooAny<T extends any> {}

class BarAny<T extends any> {}

function BazAny<T extends any>() {}

const QuuxAny = <T extends any>() => {};

interface FooUnknown<T extends unknown> {}

class BarUnknown<T extends unknown> {}

function BazUnknown<T extends unknown>() {}

const QuuxUnknown = <T extends unknown>() => {};
```

Examples of **correct** code for this rule:

```typescript
interface Foo<T> {}

class Bar<T> {}

function Baz<T>() {}

const Quux = <T>() => {};
```

## Differences from ESLint

- JSDoc `@template {any} T` / `@template {unknown} T` will be reported; the upstream ESLint rule does not report them.

## Original Documentation

- [typescript-eslint rule: no-unnecessary-type-constraint](https://typescript-eslint.io/rules/no-unnecessary-type-constraint)
