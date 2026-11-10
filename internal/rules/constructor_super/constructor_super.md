# constructor-super

## Rule Details

Verifies that constructors of derived classes (classes that extend another class) call `super()`, and that constructors of non-derived classes do not call `super()`. Also detects duplicate `super()` calls in the same constructor. A return with a value substitutes for `super()`; a bare `return;` does not. Explicit constructors in classes extending `null` must return a value, because `null` cannot be called as a constructor.

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

## Autofix and Suggestions

This rule does not provide autofixes or suggestions.

## Differences from Upstream

ESLint v10.8.1 reports only `missingAll` for the following constructor with a `for` update expression. Rslint reports `missingSome` and `duplicate`: the loop can execute zero times or call `super()` repeatedly.

```javascript
class A extends B {
  constructor() {
    for (var i = 0; i < 10; i++) {
      super();
    }
  }
}
```

Rslint also reports `badSuper` for `super()` in a constructor without an `extends` clause. ESLint's parser rejects this syntax before the rule runs.

## Original Documentation

- [ESLint: constructor-super](https://eslint.org/docs/latest/rules/constructor-super)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/constructor-super.js)
