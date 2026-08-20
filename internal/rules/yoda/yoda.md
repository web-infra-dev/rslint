# yoda

## Rule Details

Yoda conditions are so named because the literal value of the condition comes first while the variable comes second, e.g. `if ("red" === color)`. This rule enforces a consistent style of conditions which compare a variable to a literal value.

Examples of **incorrect** code for this rule, with the default `"never"` option:

```javascript
if ("red" === color) {
}

if (true == flag) {
}

if (5 > count) {
}

if (-1 < str.indexOf(substr)) {
}

if (0 <= x && x < 1) {
}
```

Examples of **correct** code for this rule, with the default `"never"` option:

```javascript
if (value === "red") {
}

if (flag == true) {
}

if (count < 5) {
}
```

## Options

This rule takes a string option:

- `"never"` (default): comparisons must never be Yoda conditions.
- `"always"`: the literal value must always come first.

The `"never"` option can take exception options in an object literal:

- `exceptRange`: when `true`, allows Yoda conditions in range comparisons that are wrapped directly in parentheses, including the parentheses of an `if` or `while` condition. A range comparison tests whether a variable is inside or outside the range between two literal values. Default `false`.
- `onlyEquality`: when `true`, only reports Yoda conditions for the equality operators `==` and `===`. Default `false`.

`onlyEquality` allows a superset of the exceptions `exceptRange` allows, so combining both options together isn't useful.

Examples of **correct** code for this rule with `{ "exceptRange": true }`:

```json
{ "yoda": ["error", "never", { "exceptRange": true }] }
```

```javascript
function isReddish(color) {
  return (color.hue < 60 || 300 < color.hue);
}

if (x < -1 || 1 < x) {
}

if ((0 <= rand && rand < 1) && count < 10) {
}
```

Each parenthesized pair of comparisons forms one range comparison, so a range comparison combined with a further condition needs its own parentheses.

Examples of **correct** code for this rule with `{ "onlyEquality": true }`:

```json
{ "yoda": ["error", "never", { "onlyEquality": true }] }
```

```javascript
if (x < -1 || 9 < x) {
}

if (x !== "foo" && "bar" != x) {
}
```

Examples of **incorrect** code for this rule with the `"always"` option:

```json
{ "yoda": ["error", "always"] }
```

```javascript
if (color == "blue") {
}
```

Examples of **correct** code for this rule with the `"always"` option:

```json
{ "yoda": ["error", "always"] }
```

```javascript
if ("blue" == color) {
}
```

## Differences from ESLint

- TypeScript-only wrappers with no runtime effect — `x!`, `x as T`, `x satisfies T` — are read through when deciding whether the two comparisons of a range test hold the same operand. With `{ "exceptRange": true }`, `if (0 <= x! && x! < 1) {}` reads as one range comparison and stays exempt.

## Original Documentation

- [ESLint: yoda](https://eslint.org/docs/latest/rules/yoda)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.0/lib/rules/yoda.js)
