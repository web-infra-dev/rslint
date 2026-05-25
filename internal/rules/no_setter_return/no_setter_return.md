# no-setter-return

## Rule Details

Disallows returning a value from a setter. Setters cannot meaningfully return values since any return value is silently ignored by the JavaScript engine. A bare `return;` (without a value) is allowed for control flow purposes.

Examples of **incorrect** code for this rule:

```javascript
var foo = {
  set a(val) {
    return 1;
  },
};

class A {
  set a(val) {
    return val;
  }
}

var bar = {
  set a(val) {
    return undefined;
  },
};
```

Examples of **correct** code for this rule:

```javascript
var foo = {
  set a(val) {
    val = 1;
  },
};

class A {
  set a(val) {
    if (!val) {
      return; // bare return for flow control is fine
    }
    this._a = val;
  }
}

class B {
  get a() {
    return this._a; // getters can return values
  }
}
```

## Original Documentation

- [ESLint no-setter-return](https://eslint.org/docs/latest/rules/no-setter-return)
