# no-unnecessary-type-arguments

## Rule Details

Disallow type arguments that are equal to the default.

If a type parameter has a default value and an explicit type argument provides the same type, the argument is redundant and can be omitted for cleaner code.

Examples of **incorrect** code for this rule:

```typescript
function f<T = number>() {}
f<number>(); // number is the default, can be omitted

type Foo<T = string> = T[];
const x: Foo<string> = []; // string is the default

class Bar<T = boolean> {}
new Bar<boolean>(); // boolean is the default
```

Examples of **correct** code for this rule:

```typescript
function f<T = number>() {}
f(); // uses default
f<string>(); // overrides default

type Foo<T = string> = T[];
const x: Foo = [];
const y: Foo<number> = [];
```

## Original Documentation

- [typescript-eslint no-unnecessary-type-arguments](https://typescript-eslint.io/rules/no-unnecessary-type-arguments)
