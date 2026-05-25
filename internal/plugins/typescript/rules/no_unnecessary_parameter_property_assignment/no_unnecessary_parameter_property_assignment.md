# no-unnecessary-parameter-property-assignment

## Rule Details

Disallow unnecessary assignment of constructor property parameter.

TypeScript's parameter property syntax (`constructor(public foo: string)`) both declares a class member and assigns the constructor argument to it. Writing an explicit `this.foo = foo` inside the constructor body therefore performs the exact same assignment a second time and adds nothing. This rule reports `this.X = X` — and the `||=`, `&&=`, `??=` variants that have the same effect on a freshly-bound member — when the constructor's parameter list declares a parameter property named `X`.

Examples of **incorrect** code for this rule:

```typescript
class Foo {
  constructor(public foo: string) {
    this.foo = foo;
  }
}

class Foo {
  constructor(public foo: string) {
    this.foo ||= foo;
  }
}

class Foo {
  constructor(public foo: string) {
    this.foo ??= foo;
  }
}

class Foo {
  constructor(public foo: string) {
    this.foo &&= foo;
  }
}

class Foo {
  constructor(private foo: string) {
    this['foo'] = foo;
  }
}

class Foo {
  constructor(public foo?: string) {
    this.foo = foo!;
  }
}

class Foo {
  constructor(public foo?: string) {
    this.foo = foo as any;
  }
}
```

Examples of **correct** code for this rule:

```typescript
class Foo {
  constructor(public foo: string) {}
}

class Foo {
  constructor(private foo: string) {
    this.foo = bar;
  }
}

class Foo {
  foo: string;
  constructor(foo: string) {
    this.foo = foo;
  }
}

class Foo {
  constructor(public foo: number) {
    this.foo += foo;
    this.foo -= foo;
  }
}

class Foo {
  constructor(public foo: number) {
    this.foo += 1;
    this.foo = foo;
  }
}

class Foo {
  constructor(public foo: number) {
    {
      const foo = 1;
      this.foo = foo;
    }
  }
}
```

## Original Documentation

- [typescript-eslint no-unnecessary-parameter-property-assignment](https://typescript-eslint.io/rules/no-unnecessary-parameter-property-assignment)
