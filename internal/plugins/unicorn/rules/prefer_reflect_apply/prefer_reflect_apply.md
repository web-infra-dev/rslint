# prefer-reflect-apply

## Rule details

Prefer `Reflect.apply()` over `Function#apply()`. This avoids relying on the target function's `apply` property, which may be missing or overridden.

The rule checks calls with `null` or `this` as the receiver and an array literal or `arguments` as the argument list. It also checks `Function.prototype.apply.call()` with these arguments. Statically known computed method names are supported, but variables used as property names are not resolved.

Examples of **incorrect** code:

```javascript
fn.apply(null, [42]);
fn.apply(this, arguments);
Function.prototype.apply.call(fn, null, [42]);
```

Examples of **correct** code:

```javascript
Reflect.apply(fn, null, [42]);
Reflect.apply(fn, this, arguments);
```

## Options

This rule has no options.

## Differences from upstream

Automatic fixes preserve parentheses around sequence expressions: `(first, second).apply(null, [])` becomes `Reflect.apply((first, second), null, [])`. Upstream v75.0.0 removes these parentheses, changing the call's arguments.

Calls such as `super.apply(null, [])` and `object?.method.apply(null, [])` are reported without a fix. Converting them would produce an invalid bare `super` argument or lose optional-chain short-circuiting. Explicit optional calls such as `fn.apply?.(null, [])` are ignored, as upstream does.

Computed keys must pass a conservative side-effect check before a fix removes them. For example, `fn[(log(), "apply")](null, [])` and `fn[key = "apply"](null, [])` are reported without a fix so the call or assignment is not lost. This check applies to every removed key in `Function.prototype.apply.call()` too. Calls, assignments, potentially unsafe property reads, unknown references and unsafe coercions prevent a fix, even on a statically unreachable branch. Class/JSX creation, tagged templates and spread expressions are also left unchanged because they can execute code implicitly. Effects in the retained target or argument list, such as `getFunction().apply(null, [effect()])`, do not prevent a fix.

Calls with a spread as the first argument to `Function.prototype.apply.call()` are also reported without a fix. For example, when `targets` is `[fn, null]`, `Function.prototype.apply.call(...targets, null, [])` invokes `fn` without arguments, while the corresponding `Reflect.apply()` call would throw. The rule does not try to infer the spread's length.

## Original documentation

- [eslint-plugin-unicorn: prefer-reflect-apply](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/prefer-reflect-apply.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/prefer-reflect-apply.js)
