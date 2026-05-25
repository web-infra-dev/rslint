# no-unnecessary-qualifier

## Rule Details

Disallow unnecessary namespace qualifiers. Accessing an enum member or a namespace export by its qualified name when the right-hand identifier is already directly in scope is redundant. The rule flags those qualifiers and offers an autofix that removes them.

Examples of **incorrect** code for this rule:

```typescript
enum A {
  B,
  C = A.B,
}

namespace A {
  export type B = number;
  const x: A.B = 3;
}

namespace A {
  export namespace B {
    export type T = number;
    const x: A.B.T = 3;
  }
}
```

Examples of **correct** code for this rule:

```typescript
enum A {
  B,
  C = B,
}

namespace A {
  export type B = number;
  const x: B = 3;
}

namespace X {
  export type T = number;
}

namespace Y {
  export const x: X.T = 3;
}
```

## Original Documentation

- [typescript-eslint no-unnecessary-qualifier](https://typescript-eslint.io/rules/no-unnecessary-qualifier)
