# no-this-outside-of-class

## Rule Details

Disallow `this` outside of classes. `this` should only be used when JavaScript
class syntax or an explicit TypeScript `this` parameter defines the receiver.

Examples of **incorrect** code for this rule:

```javascript
function Foo(value) {
  this.value = value;
}

const foo = {
  method() {
    return this.value;
  },
};

Foo.prototype.method = function () {
  this.value();
};

class Foo {
  method() {
    function getValue() {
      return this.value;
    }

    return getValue();
  }
}
```

Examples of **correct** code for this rule:

```javascript
class Foo {
  constructor(value) {
    this.value = value;
  }

  method() {
    return this.value;
  }
}

class Foo {
  method() {
    const getValue = () => this.value;
    return getValue();
  }
}

const foo = {
  validator(this: TrackedModel) {
    return this.value;
  },
};
```

## Original Documentation

- [eslint-plugin-unicorn: no-this-outside-of-class](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v72.0.0/docs/rules/no-this-outside-of-class.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v72.0.0/rules/no-this-outside-of-class.js)
