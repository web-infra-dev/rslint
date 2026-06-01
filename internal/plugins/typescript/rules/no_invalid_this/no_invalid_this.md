# no-invalid-this

Disallow `this` keywords outside of classes or class-like objects.

## Rule Details

Under strict mode, `this` keywords outside of classes or class-like objects might be `undefined` and raise a `TypeError`.

This rule judges from the following conditions whether or not the function is a constructor:

- The name of the function starts with uppercase.
- The function is assigned to a variable which starts with an uppercase letter.
- The function is a constructor of ES2015 Classes.

This rule judges from the following conditions whether or not the function is a method:

- The function is on an object literal.
- The function is assigned to a property.
- The function is a method / getter / setter of ES2015 Classes.

And this rule allows `this` keywords in functions below:

- The `call` / `apply` / `bind` method of the function is called directly.
- The function is a callback of array methods (such as `.forEach()`) if `thisArg` is given.
- The function has an `@this` tag in its JSDoc comment.
- The function declares an explicit `this` parameter (`function foo(this: SomeType)`).

Otherwise, this rule warns on `this` keywords. It also reports `this` at the top level.

Examples of **incorrect** code for this rule:

```typescript
this.a = 0;
baz(() => this);

(function () {
  this.a = 0;
  baz(() => this);
})();

function foo() {
  this.a = 0;
  baz(() => this);
}

var foo = function () {
  this.a = 0;
  baz(() => this);
};

foo(function () {
  this.a = 0;
  baz(() => this);
});

obj.foo = () => {
  // `this` of arrow functions is the outer scope's.
  this.a = 0;
};

var obj = {
  aaa: function () {
    return function foo() {
      // There is a method `aaa`, but `foo` is not a method.
      this.a = 0;
      baz(() => this);
    };
  },
};

foo.forEach(function () {
  this.a = 0;
  baz(() => this);
});
```

Examples of **correct** code for this rule:

```typescript
function Foo() {
  // OK, legacy-style constructor.
  this.a = 0;
  baz(() => this);
}

class Foo {
  constructor() {
    this.a = 0;
    baz(() => this);
  }
}

var obj = {
  foo() {
    this.a = 0;
  },
};

var obj = {
  get foo() {
    return this.a;
  },
};

Object.defineProperty(obj, 'foo', {
  value: function foo() {
    this.a = 0;
  },
});

obj.foo = function foo() {
  this.a = 0;
};

class Foo {
  foo() {
    this.a = 0;
    baz(() => this);
  }

  static foo() {
    this.a = 0;
    baz(() => this);
  }
}

var foo = function foo() {
  this.a = 0;
}.bind(obj);

foo.forEach(function () {
  this.a = 0;
  baz(() => this);
}, thisArg);

/** @this Foo */
function foo() {
  this.a = 0;
}

function foo(this: SomeType) {
  this.a = 0;
}
```

## Options

### `capIsConstructor`

**Type:** `boolean` — **Default:** `true`

When `true`, the rule treats a function whose name starts with an uppercase letter (or which is assigned to such a variable) as an ES5 constructor — `this` inside is allowed. Set this option to `false` to treat capitalized-name functions as regular functions.

Examples of **incorrect** code with `{ "capIsConstructor": false }`:

```json
{ "@typescript-eslint/no-invalid-this": ["error", { "capIsConstructor": false }] }
```

```typescript
function Foo() {
  this.a = 0;
}

var Bar = function () {
  this.a = 0;
};

Baz = function () {
  this.a = 0;
};
```

Examples of **correct** code with `{ "capIsConstructor": false }`:

```json
{ "@typescript-eslint/no-invalid-this": ["error", { "capIsConstructor": false }] }
```

```typescript
obj.Foo = function () {
  // OK, assigned to a property.
  this.a = 0;
};

class Foo {
  constructor() {
    this.a = 0;
  }
}
```

## When Not To Use It

If you do not want to be notified about usage of the `this` keyword outside of classes or class-like objects, you can safely disable this rule.

## Original Documentation

- [typescript-eslint no-invalid-this](https://typescript-eslint.io/rules/no-invalid-this)
- [ESLint no-invalid-this](https://eslint.org/docs/latest/rules/no-invalid-this)
