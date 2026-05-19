# no-unsafe-function-type

## Rule Details

Disallow using the built-in `Function` type. `Function` describes any callable value: it accepts any number of arguments, returns `any`, and includes class declarations, which throw at runtime when invoked without `new`. A concrete signature — including parameters and return type — should be used instead.

A local declaration named `Function` (a `type` alias, `interface`, or `class`) shadows the global one, in which case the reference is no longer the unsafe built-in and is not reported.

Examples of **incorrect** code for this rule:

```typescript
let value: Function;

let values: Function[];

let valueOrNumber: Function | number;

class Weird implements Function {
  // ...
}

interface AlsoWeird extends Function {
  // ...
}
```

Examples of **correct** code for this rule:

```typescript
let value: () => void;

let value: <T>(t: T) => T;

{
  type Function = () => void;
  let value: Function;
}
```

## Original Documentation

- [typescript-eslint no-unsafe-function-type](https://typescript-eslint.io/rules/no-unsafe-function-type)
