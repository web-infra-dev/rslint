# no-unsafe-optional-chaining

## Rule Details

Disallows using optional chaining in contexts where the `undefined` value is not allowed. Optional chaining (`?.`) can short-circuit to `undefined`. When the result is used in a position where `undefined` causes a TypeError or unexpected behavior, this rule reports it.

Examples of **incorrect** code for this rule:

```javascript
(obj?.foo)(); // TypeError if obj?.foo is undefined
(obj?.foo).bar; // TypeError
new (obj?.foo)(); // TypeError
const { a } = obj?.foo; // TypeError (destructuring undefined)
[...obj?.foo]; // TypeError (spreading undefined)
for (const x of obj?.foo) {
} // TypeError (iterating undefined)
'foo' in obj?.bar; // TypeError
foo instanceof obj?.bar; // TypeError
class Foo extends obj?.bar {} // TypeError
```

Examples of **correct** code for this rule:

```javascript
obj?.foo; // standalone is fine
obj?.foo(); // optional call is fine
(obj?.foo ?? bar)(); // fallback via ??
(obj?.foo || bar).baz; // fallback via ||
obj?.foo?.bar; // chaining is fine
```

## Options

### `disallowArithmeticOperators`

When set to `true`, also reports arithmetic operations on optional chaining results, which can produce `NaN`. Default is `false`.

Examples of **incorrect** code with `{ "disallowArithmeticOperators": true }`:

```javascript
obj?.foo + bar; // may be NaN
+obj?.foo; // may be NaN
-obj?.foo; // may be NaN
obj?.foo * bar; // may be NaN
```

## Original Documentation

- [ESLint no-unsafe-optional-chaining](https://eslint.org/docs/latest/rules/no-unsafe-optional-chaining)
