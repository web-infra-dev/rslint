# no-unused-private-class-members

## Rule Details

Disallow unused private class members.

This rule extends the base ESLint [`no-unused-private-class-members`](https://eslint.org/docs/latest/rules/no-unused-private-class-members) rule by recognizing members declared with TypeScript's `private` accessibility modifier and parameter properties (`constructor(private x: T)`), in addition to JavaScript's `#`-prefixed private fields.

A private class member is considered *used* when its value is read at least once. Pure writes do not count — a field that is only assigned but never read can be removed with no observable effect. Getter and setter accessors are the exception: any access keeps them alive, since they can have side effects.

Examples of **incorrect** code for this rule:

```typescript
class A {
  #foo = 123;
}
```

```typescript
class A {
  private foo = 123;
}
```

```typescript
class A {
  private foo = 123;
  bar() {
    this.foo = 1;
    this.foo += 2;
    this.foo++;
  }
}
```

```typescript
class A {
  constructor(private foo: number) {}
}
```

Examples of **correct** code for this rule:

```typescript
class A {
  #foo = 123;
  bar() {
    return this.#foo;
  }
}
```

```typescript
class A {
  private foo = 123;
  constructor() {
    console.log(this.foo);
  }
}
```

```typescript
class A {
  private foo: number = 0;
  bar(other: A) {
    return other.foo;
  }
}
```

```typescript
class A {
  private accessor foo = 123;
  bar() {
    this.foo = 0;
  }
}
```

## Options

This rule has no options.

## Known Limitations

Detection is shape-based, so the rule cannot see uses that reach the member through these indirect patterns:

- Access through a variable whose type annotation is more complex than a single `T` or `typeof T` (unions, intersections, generic constraints).
- External access via bracket notation with a dynamic key (`instance[someVar]`).
- Usages reached through multi-step `this`-aliasing (`let X = this; let Y = X; Y.foo`).

## Original Documentation

- [typescript-eslint no-unused-private-class-members](https://typescript-eslint.io/rules/no-unused-private-class-members)
- [ESLint no-unused-private-class-members](https://eslint.org/docs/latest/rules/no-unused-private-class-members)
