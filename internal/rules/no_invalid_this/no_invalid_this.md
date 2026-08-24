# no-invalid-this

## Rule Details

Under strict mode, `this` keywords outside of classes or class-like objects might be `undefined` and raise a `TypeError`. This rule applies **only** in strict mode; sloppy-mode code (a plain script with no `"use strict"` directive and no ES module syntax) is never flagged.

Top-level `this` at the top of a script always refers to the global object and is valid. Top-level `this` in an ES module is always invalid, since its value is `undefined`.

For `this` inside functions, this rule judges from the following conditions whether or not the function is a constructor:

- The name of the function starts with uppercase.
- The function is assigned to a variable which starts with an uppercase letter.
- The function is a constructor of ES2015 classes.

This rule judges from the following conditions whether or not the function is a method:

- The function is on an object literal.
- The function is assigned to a property.
- The function is a method / getter / setter of ES2015 classes.

And this rule allows `this` keywords in functions below:

- The `call` / `apply` / `bind` method of the function is called directly.
- The function is a callback of array methods (such as `.forEach()`) if `thisArg` is given.
- The function has an `@this` tag in its JSDoc comment.
- The function declares an explicit `this` parameter (`function foo(this: SomeType)`).

And this rule always allows `this` keywords in the following contexts:

- At the top level of scripts.
- In class field initializers.
- In class static blocks.

Otherwise, this rule warns on `this` keywords.

Examples of **incorrect** code for this rule in strict mode:

```typescript
'use strict';

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

Examples of **correct** code for this rule in strict mode:

```typescript
'use strict';

this.a = 0;
baz(() => this);

function Foo() {
  // OK, this is in a legacy style constructor.
  this.a = 0;
  baz(() => this);
}

class Foo {
  constructor() {
    // OK, this is in a constructor.
    this.a = 0;
    baz(() => this);
  }
}

var obj = {
  foo() {
    // OK, this is in a method (this function is on an object literal).
    this.a = 0;
  },
};

var obj = {
  get foo() {
    // OK, this is in a method (this function is on an object literal).
    return this.a;
  },
};

Object.defineProperty(obj, 'foo', {
  value: function foo() {
    // OK, this is in a method (this function is on an object literal).
    this.a = 0;
  },
});

obj.foo = function foo() {
  // OK, this is in a method (this function assigns to a property).
  this.a = 0;
};

class Baz {
  // OK, this is in a class field initializer.
  a = this.b;

  // OK, static initializers also have valid this.
  static a = this.b;

  foo() {
    // OK, this is in a method.
    this.a = 0;
    baz(() => this);
  }

  static foo() {
    // OK, this is in a method (static methods also have valid this).
    this.a = 0;
    baz(() => this);
  }

  static {
    // OK, static blocks also have valid this.
    this.a = 0;
    baz(() => this);
  }
}

var foo = function foo() {
  // OK, the bind method of this function is called directly.
  this.a = 0;
}.bind(obj);

foo.forEach(function () {
  // OK, thisArg of .forEach() is given.
  this.a = 0;
  baz(() => this);
}, thisArg);

/** @this Foo */
function foo() {
  // OK, this function has a @this tag in its JSDoc comment.
  this.a = 0;
}

function foo(this: SomeType) {
  // OK, this function has an explicit `this` parameter.
  this.a = 0;
}
```

## Options

This rule has an object option, with one option:

- `"capIsConstructor": false` (default `true`) disables the assumption that a function whose name starts with an uppercase letter is a constructor.

### `capIsConstructor`

By default, this rule always allows the use of `this` in functions whose name starts with an uppercase letter and anonymous functions that are assigned to a variable whose name starts with an uppercase letter, assuming that those functions are used as constructor functions.

Set `"capIsConstructor"` to `false` if you want those functions to be treated as regular functions.

Examples of **incorrect** code for this rule with `{ "capIsConstructor": false }`:

```json
{ "no-invalid-this": ["error", { "capIsConstructor": false }] }
```

```typescript
'use strict';

function Foo() {
  this.a = 0;
}

var Bar = function Foo() {
  this.a = 0;
};

Baz = function () {
  this.a = 0;
};
```

Examples of **correct** code for this rule with `{ "capIsConstructor": false }`:

```json
{ "no-invalid-this": ["error", { "capIsConstructor": false }] }
```

```typescript
'use strict';

obj.Foo = function Foo() {
  // OK, this is in a method.
  this.a = 0;
};
```

## Differences from ESLint

- rslint determines whether code is an ES module purely from the presence of `import` / `export` syntax in the file. ESLint additionally lets `languageOptions.sourceType` and `parserOptions.ecmaFeatures.globalReturn` override this independently of file content; rslint does not expose either setting, so top-level `this` validity always follows the file's actual module-ness.

## When Not To Use It

If you do not want to be notified about usage of the `this` keyword outside of classes or class-like objects, you can safely disable this rule.

## Original Documentation

- [ESLint: no-invalid-this](https://eslint.org/docs/latest/rules/no-invalid-this)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-invalid-this.js)
