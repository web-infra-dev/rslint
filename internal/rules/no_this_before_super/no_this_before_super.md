# no-this-before-super

## Rule Details

Disallows use of `this` or `super` (for property access like `super.foo`) before calling `super()` in constructors of derived classes.

In a derived class (a class that extends another class), the constructor must call `super()` before accessing `this` or `super` for property access. Accessing `this` or `super` before `super()` has been called will throw a `ReferenceError` at runtime.

Examples of **incorrect** code for this rule:

```javascript
class A extends B {
  constructor() {
    this.a = 0; // "this" before "super()"
    super();
  }
}

class A extends B {
  constructor() {
    super.foo(); // "super" property access before "super()"
    super();
  }
}

class A extends B {
  constructor() {
    super(this.a); // "this" in super() arguments
  }
}
```

Examples of **correct** code for this rule:

```javascript
class A extends B {
  constructor() {
    super();
    this.a = 0;
  }
}

class A {
  constructor() {
    this.a = 0; // OK - not a derived class
  }
}

class A extends B {
  constructor() {
    super();
    super.foo();
  }
}
```

## Original Documentation

- [ESLint no-this-before-super](https://eslint.org/docs/latest/rules/no-this-before-super)
