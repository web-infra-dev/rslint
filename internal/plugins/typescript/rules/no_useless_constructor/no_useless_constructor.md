# no-useless-constructor

## Rule Details

Disallow unnecessary constructors.

ES2015 provides a default class constructor if one is not specified. As such, it is unnecessary to provide an empty constructor or one that simply delegates into its parent class, as in the following examples:

This rule extends ESLint's `no-useless-constructor` with TypeScript-specific support for:
- Access modifiers (`private`, `protected`, `public`)
- Parameter properties
- Parameter decorators

Examples of **incorrect** code for this rule:

```typescript
class A {
  constructor() {}
}

class B extends A {
  constructor(foo) {
    super(foo);
  }
}

class C {
  public constructor() {}
}
```

Examples of **correct** code for this rule:

```typescript
class A {
  constructor(private name: string) {}
}

class B {
  private constructor() {}
}

class C extends D {
  public constructor(foo) {
    super(foo);
  }
}
```

## Original Documentation

- [typescript-eslint no-useless-constructor](https://typescript-eslint.io/rules/no-useless-constructor)
- [ESLint no-useless-constructor](https://eslint.org/docs/latest/rules/no-useless-constructor)
