# class-methods-use-this

## Rule Details

Enforces that class methods use `this` (or `super`) to avoid unintentional reliance on `this` binding.

Examples of **incorrect** code for this rule:

```ts
class Foo {
  method() {}
}
```

```ts
class Foo {
  property = () => {};
}
```

Examples of **correct** code for this rule:

```ts
class Foo {
  method() {
    this.value = 1;
  }
}
```

```ts
class Foo {
  property = () => {
    this.value = 1;
  };
}
```

## Original Documentation

https://typescript-eslint.io/rules/class-methods-use-this
