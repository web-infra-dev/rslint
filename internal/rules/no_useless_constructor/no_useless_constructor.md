# no-useless-constructor

## Rule Details

Disallow unnecessary constructors.

ES2015 provides a default class constructor if one is not specified. As such, it is unnecessary to provide an empty constructor or one that simply delegates into its parent class.

Examples of **incorrect** code for this rule:

```javascript
class A {
  constructor() {}
}

class B extends A {
  constructor(...args) {
    super(...args);
  }
}
```

Examples of **correct** code for this rule:

```javascript
class A {}

class B {
  constructor() {
    doSomething();
  }
}

class C extends A {
  constructor() {
    super('foo');
  }
}

class D extends A {
  constructor() {
    super();
    doSomething();
  }
}
```

## Original Documentation

- [ESLint no-useless-constructor](https://eslint.org/docs/latest/rules/no-useless-constructor)
