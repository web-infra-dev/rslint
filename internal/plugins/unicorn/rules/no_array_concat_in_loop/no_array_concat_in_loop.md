# no-array-concat-in-loop

## Rule Details

Disallows accumulating an array with `Array#concat()` inside a loop. Because
`concat()` creates a new array and copies the values accumulated so far, using
it on every iteration can make the loop quadratic in the total number of items.

This rule reports local `let` and `var` variables initialized to an empty
array literal. It does not try to infer arbitrary array-like values or custom
`concat()` methods.

Examples of **incorrect** code for this rule:

```javascript
let result = [];

for (const chunk of chunks) {
  result = result.concat(chunk);
}
```

Examples of **correct** code for this rule:

```javascript
const result = [];

for (const chunk of chunks) {
  result.push(...chunk);
}
```

```javascript
const result = chunks.flat();
```

```javascript
let result = [];

for (const chunk of chunks) {
  result = other.concat(chunk);
}
```

The rule does not provide an autofix. Replacing `concat()` with `push()` can
change observable behavior when the previous array is aliased, and the two
methods have different argument-spreading semantics.

## Original Documentation

- [eslint-plugin-unicorn: no-array-concat-in-loop](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/no-array-concat-in-loop.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/no-array-concat-in-loop.js)
