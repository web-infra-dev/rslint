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
Explicit TypeScript and JSDoc return annotations on local factories are respected
when identifying a non-array receiver.

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
When necessary, a semicolon keeps the following statement outside the loop.
Comments in the removed text and multiline tokens prevent automatic fixes.
Loops with an `else` branch, multiple statements in the guarded block, or
`for await…of` are ignored. A loop binding referenced by the iterable also
prevents a report.

## Options

This rule has no options.

## Differences from upstream

Property reads in evaluated call arguments or callbacks offer suggestions
because a getter can run user code or change the array's length. For example,
`array.some(value => value.ok) && array.length > 0` keeps its length guard until
the suggestion is accepted. Unicorn v79.0.0 automatically removes this guard.

For the following leading check, rslint offers an editor suggestion because
evaluating the tagged template can run user code even when the array is empty.
Unicorn v79.0.0 removes this check automatically.

```javascript
array.length > 0 && array.some(tag`text`);
```

Applying the suggestion calls `tag` even when `array` is empty. Review whether
that call is safe before accepting the suggestion.

Leading checks also offer suggestions when evaluating the method arguments
performs a spread, including spreads nested inside arrays, objects, or JSX:

```javascript
array.length > 0 && array.some(...callbacks);
array.length === 0 || array.every(predicate, { ...options });
```

Removing the guard can consume an iterator or invoke getters even when the
array is empty. For leading checks, spreads inside an inline callback's body
do not prevent an automatic fix because creating the callback does not run its
body. For trailing checks, a callback that spreads JSX props offers a suggestion
because a getter can change the array during iteration.

rslint also adds a semicolon when unwrapping a loop would attach the following
statement to its body. A following `[value].forEach(...)`, for example, stays
outside the loop instead of becoming part of the loop body's last call.

## Original Documentation

- [eslint-plugin-unicorn: no-useless-length-check](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-length-check.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-useless-length-check.js)
