# no-static-only-class

## Rule Details

A class with only static members has no instance behavior — it's effectively
just a namespace for a bag of values and functions. Use a plain object instead.

This rule reports any `class` declaration or expression where every member is
a public static field or method. Classes that extend another class, that have
private (`#name`) members, that carry class-level decorators, that contain a
`static {}` initialization block, or that mix in any non-static member are
left alone.

Examples of **incorrect** code for this rule:

```javascript
class StaticOnly {
  static foo = 1;
  static bar() {}
}

const Counter = class {
  static value = 0;
  static increment() {
    Counter.value++;
  }
};
```

Examples of **correct** code for this rule:

```javascript
const StaticOnly = {
  foo: 1,
  bar() {},
};

// Has a non-static member — leave it as a class.
class WithInstance {
  static greeting = "hi";
  log() {
    console.log(WithInstance.greeting);
  }
}

// Extends another class — the static members might supplement inherited
// instance behavior.
class Sub extends Base {
  static helper() {}
}

// Private members force the class form.
class PrivateOnly {
  static #secret = 42;
  static reveal() {
    return PrivateOnly.#secret;
  }
}

// Static initialization blocks keep the class form.
class WithStaticBlock {
  static {
    /* setup */
  }
  static run() {}
}
```

## Differences from ESLint

The autofix is suppressed in a few TypeScript-only cases where ESLint's
algorithm would produce syntactically invalid TypeScript. The diagnostic
is still reported in all of them.

- Class with type parameters (`class A<T> { ... }`). ESLint's autofix
  yields `const A<T> = { ... }`; `const` declarations do not accept type
  parameters.
- Property with the optional postfix `?` (`static a?`). ESLint's autofix
  yields `{ a?: ... }`; object members cannot be declared optional
  (TS1162).
- Property with the definite-assignment assertion postfix `!`
  (`static a! = 1`). ESLint's autofix yields `{ a! : ... }`; the `!`
  assertion is not permitted in this context (TS1255).
- Property declared with the TC39 `accessor` keyword
  (`static accessor a = 1`). The keyword has no object-literal analog.
- Method-like member (method, getter, setter, or constructor) declared
  without a body, e.g. an overload signature
  (`static a(x: number): void;`). Object-literal methods must have a
  body.

## Original Documentation

- ESLint plugin documentation: <https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/no-static-only-class.md>
- Source code: <https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/rules/no-static-only-class.js>
