# constructor-super

## Rule Details

Verifies that constructors of derived classes (classes that extend another class) call `super()`, and that constructors of non-derived classes do not call `super()`. Also detects duplicate `super()` calls in the same constructor and ensures `super()` is called in all code paths.

Examples of **incorrect** code for this rule:

```javascript
class A extends B {
  constructor() {
    // missing super() call
  }
}

class A {
  constructor() {
    super(); // super() in non-derived class
  }
}

class A extends B {
  constructor() {
    super();
    super(); // duplicate super() call
  }
}
```

Examples of **correct** code for this rule:

```javascript
class A extends B {
  constructor() {
    super();
  }
}

class A {
  constructor() {
    // no super() needed
  }
}

class A extends B {
  constructor(cond) {
    if (cond) {
      super(true);
    } else {
      super(false);
    }
  }
}
```

## Original Documentation

- [ESLint constructor-super](https://eslint.org/docs/latest/rules/constructor-super)
