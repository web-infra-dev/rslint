# no-constructor-return

## Rule Details

Disallows `return` statements that return a value from class constructors. Returning a value from a constructor of a class is usually a mistake, as the returned value is only used when the constructor is called without `new`. Bare `return` statements (without a value) are allowed for flow control purposes.

Examples of **incorrect** code for this rule:

```javascript
class A {
  constructor() {
    return 'value';
  }
}

class B {
  constructor() {
    return { something: true };
  }
}
```

Examples of **correct** code for this rule:

```javascript
class A {
  constructor() {
    this.value = 42;
  }
}

class B {
  constructor() {
    if (!valid) {
      return; // bare return for flow control is fine
    }
    this.init();
  }
}
```

## Original Documentation

- [ESLint no-constructor-return](https://eslint.org/docs/latest/rules/no-constructor-return)
