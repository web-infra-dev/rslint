# getter-return

## Rule Details

Enforces that property getters contain a `return` statement that returns a value. This applies to getter methods in object literals, class declarations, and property descriptors passed to `Object.defineProperty`, `Object.defineProperties`, `Reflect.defineProperty`, and `Object.create`.

Examples of **incorrect** code for this rule:

```javascript
var obj = {
  get name() {
    // no return
  },
};

class Foo {
  get bar() {
    // no return
  }
}

Object.defineProperty(obj, 'prop', {
  get: function () {
    // no return
  },
});
```

Examples of **correct** code for this rule:

```javascript
var obj = {
  get name() {
    return 'foo';
  },
};

class Foo {
  get bar() {
    return this._bar;
  }
}

Object.defineProperty(obj, 'prop', {
  get: function () {
    return this._prop;
  },
});
```

## Original Documentation

- [ESLint getter-return](https://eslint.org/docs/latest/rules/getter-return)
