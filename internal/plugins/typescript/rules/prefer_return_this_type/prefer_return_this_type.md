# prefer-return-this-type

## Rule Details

Enforce that `this` is used when only `this` type is returned. If a class method's return type is the class itself and the method only returns `this`, then it is better to use `this` as the return type. Using the `this` return type enables proper type narrowing in subclasses, since `this` refers to the current class type rather than a fixed class name.

Examples of **incorrect** code for this rule:

```typescript
class Foo {
  doStuff(): Foo {
    return this;
  }
  chain(): Foo {
    return this;
  }
}
```

Examples of **correct** code for this rule:

```typescript
class Foo {
  doStuff(): this {
    return this;
  }
  chain(): this {
    return this;
  }
}
```

## Original Documentation

- [typescript-eslint prefer-return-this-type](https://typescript-eslint.io/rules/prefer-return-this-type)
