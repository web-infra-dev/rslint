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

The upstream reference is ESLint v10.8.1. Rslint preserves its diagnostic messages and ranges, but corrects the following code-path bugs:

- `return super()` is a single call, not a duplicate. Returning a value is tracked separately from initializing `this`, so a `finally` block can replace the pending return.
- A literal-false loop does not revisit its body. `do { super(); } while (false)` calls once; pure constant loop conditions and parentheses do not introduce imaginary exits or back edges.
- Every label directly attached to a loop refers to that loop. `outer: inner: while (true) { super(); continue outer; }` has a repeatable call and reports `duplicate`.
- Binding defaults use the incoming value: a known own data property with `undefined` triggers the default, a known supplied value skips it, and unknown values retain both paths. Getter, iterator and assignment operations remain possible exception sources.
- A catch is reached from potentially throwing operations, not declaration names or the normal end of a try. Retrying a `super()` that threw before binding `this` is not itself a duplicate. Instance element initialization can throw after binding `this`; a second call in that catch is still a duplicate.
- `finally` runs for normal, returned, thrown, break and continue completions. Its own abrupt completion replaces the pending one. An unconditional object return from finally satisfies the constructor; a bare return that replaces an object return still requires initialized `this`.
- A `for` update expression does not erase the body's call state. For example, the loop below reports `missingSome` and `duplicate`; ESLint v10.8.1 reports only `missingAll`.

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

As in the pinned rule, arbitrary call return values are not inferred while selecting logical branches. The complete upstream suite retains that conservative contract.

## Original Documentation

- [ESLint: constructor-super](https://eslint.org/docs/latest/rules/constructor-super)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/constructor-super.js)
