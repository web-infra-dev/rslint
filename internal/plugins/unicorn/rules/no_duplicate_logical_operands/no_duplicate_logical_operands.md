# no-duplicate-logical-operands

## Rule Details

Disallow adjacent duplicate operands in logical expressions.

Using the same reference on adjacent sides of `&&` or `||` is usually a copy-paste mistake or redundant code. `foo && foo` and `foo || foo` evaluate the same reference twice and can be simplified to `foo`.

The rule only checks adjacent operands. It does not check non-adjacent repeats such as `foo && bar && foo`, self-comparisons such as `foo > foo`, or constant short-circuit expressions such as `true && foo`.

Examples of **incorrect** code:

```javascript
if (isReady && isReady) {
}

const result = value || value;

if (object.isReady && object.isReady) {
}
```

Examples of **correct** code:

```javascript
if (isReady) {
}

const result = value;

if (object.isReady) {
}
```

## Fixes

Duplicate identifiers and `this` expressions are autofixed.

Member expressions, and operands wrapped in TypeScript `as`, `satisfies`, `<Type>` or `!` expressions, are reported with an editor suggestion instead of an autofix because property reads can trigger getters or proxy traps.

No fix or suggestion is offered when the logical expression contains a comment outside the retained left operand.

No fix or suggestion is offered either when the logical expression is called, used as a template tag, or deleted, such as `(object.method || object.method)()`. There the expression is a plain value, while the retained operand would be a reference that changes `this`, direct `eval` or `delete` behavior. A `new` target is still fixed.

The rule skips calls, optional chains, assignments, updates, complex computed keys, and code inside `with` statements because removing one side could change runtime behavior.

## Original Documentation

- [eslint-plugin-unicorn: no-duplicate-logical-operands](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/docs/rules/no-duplicate-logical-operands.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/rules/no-duplicate-logical-operands.js)
