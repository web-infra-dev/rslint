# no-useless-length-check

## Rule Details

Disallows redundant array length checks next to `Array#some()`, `Array#every()`,
or before a `for…of` loop over the same receiver. `some()` already returns
`false` for an empty array, `every()` returns `true`, and `for…of` performs no
iterations.

The rule checks `.length === 0`, `.length !== 0`, and `.length > 0`, using the
literal spelling `0`. Use it with `unicorn/explicit-length-check` to normalize
other length checks. Known non-array receivers are ignored; typed arrays are
supported. Only adjacent conditions in the same `&&` or `||` chain match.

Examples of **incorrect** code:

```javascript
const hasMatch = array.length > 0 && array.some(predicate);
const allMatch = array.length === 0 || array.every(predicate);

if (array.length !== 0) {
  for (const element of array) {
    use(element);
  }
}
```

Examples of **correct** code:

```javascript
const hasMatch = array.some(predicate);
const allMatch = array.every(predicate);

for (const element of array) {
  use(element);
}

const nonEmptyAndAllMatch = array.length > 0 && array.every(predicate);
const emptyOrHasMatch = array.length === 0 || array.some(predicate);
```

A leading length check is automatically removed when evaluating the method
arguments has no side effects. A trailing check is automatically removed only
when the callback is an inline function without side effects. Other matching
method checks offer an editor suggestion because removing the check can change
which side effects run.

A guarded `for…of` loop is unwrapped when its indentation can be preserved.
Comments in the removed text and multiline tokens prevent automatic fixes.
Loops with an `else` branch, multiple statements in the guarded block, or
`for await…of` are ignored. A loop binding referenced by the iterable also
prevents a report.

## Options

This rule has no options.

## Differences from upstream

For the following leading check, rslint offers an editor suggestion because
evaluating the tagged template can run user code even when the array is empty.
Unicorn v79.0.0 removes this check automatically.

```javascript
array.length > 0 && array.some(tag`text`);
```

Applying the suggestion calls `tag` even when `array` is empty. Review whether
that call is safe before accepting the suggestion.

## Original Documentation

- [eslint-plugin-unicorn: no-useless-length-check](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-length-check.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-useless-length-check.js)
