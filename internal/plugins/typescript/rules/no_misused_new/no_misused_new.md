# no-misused-new

## Rule Details

Disallows incorrect usage of `new` and `constructor` in interfaces and classes. Interfaces define the shape of objects but cannot be constructed directly; defining a `new()` construct signature in an interface that returns the interface type or a `constructor` method signature is almost always a mistake. Similarly, classes should not have a method literally named `new` that returns the class type, as the `constructor` keyword should be used instead.

Examples of **incorrect** code for this rule:

```typescript
interface Foo {
  new (): Foo;
}

interface Bar {
  constructor(): Bar;
}

class Baz {
  new(): Baz;
}
```

Examples of **correct** code for this rule:

```typescript
class Foo {
  constructor() {}
}

interface Bar {
  new (): SomeOtherClass;
}

interface Baz {
  method(): void;
}
```

## Original Documentation

- [typescript-eslint no-misused-new](https://typescript-eslint.io/rules/no-misused-new)
