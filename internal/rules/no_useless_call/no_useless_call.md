# no-useless-call

## Rule Details

`Function.prototype.call()` and `Function.prototype.apply()` are slower than the
normal function invocation. This rule reports cases where `.call()` /
`.apply()` does not change `this` (so the call could be a normal invocation).

A call is reported when its `thisArg` matches the receiver implied by the
applied expression:

- `foo.call(undefined, …)` / `foo.apply(null, …)` — the applied expression has
  no implied `this`, so `null`/`undefined`/`void 0` are equivalent to a plain
  call.
- `obj.foo.call(obj, …)` / `obj.foo.apply(obj, …)` — the `thisArg` is the same
  expression (token-for-token) as the receiver of the applied member access.

`.apply()` is only flagged when the second argument is an array literal — the
variadic / spread case is the responsibility of `prefer-spread`.

Examples of **incorrect** code for this rule:

```javascript
foo.call(undefined, 1, 2, 3);
foo.apply(undefined, [1, 2, 3]);
foo.call(null, 1, 2, 3);
foo.apply(null, [1, 2, 3]);

obj.foo.call(obj, 1, 2, 3);
obj.foo.apply(obj, [1, 2, 3]);
```

Examples of **correct** code for this rule:

```javascript
// The `this` binding actually changes.
foo.call(obj, 1, 2, 3);
foo.apply(obj, [1, 2, 3]);
obj.foo.call(null, 1, 2, 3);
obj.foo.apply(null, [1, 2, 3]);
obj.foo.call(otherObj, 1, 2, 3);
obj.foo.apply(otherObj, [1, 2, 3]);

// Variadic is delegated to `prefer-spread`.
foo.apply(undefined, args);
foo.apply(null, args);
obj.foo.apply(obj, args);
```

## Known Limitations

This rule compares the applied receiver and `thisArg` at the token level
and does not model runtime side effects. Two source-identical expressions
may produce different runtime values — most notably when they contain
mutating operators. For example, `a[i++].foo.call(a[i++], 1, 2, 3)` is
flagged as unnecessary, but rewriting it to `a[i++].foo(1, 2, 3)` would
change behavior because `i` is incremented twice in the original and only
once after the rewrite. Apply the suggested rewrite with care in such
cases.

## Original Documentation

- https://eslint.org/docs/latest/rules/no-useless-call
