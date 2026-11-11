# prefer-iterator-zip

## Rule details

Prefer `Iterator.zip()` when a loop uses an index to read corresponding values
from two or more arrays and stops when the shortest array ends.

The rule checks block-bodied `for` loops with a `let` index initialized to `0`
and incremented by one. It recognizes bounds written with `Math.min()` or
conjoined length comparisons. The `Math.min()` result can also be stored in the
loop initializer or in an immediately preceding `let` or `const` declaration
used only by the condition.

Each input must be a distinct identifier. Within the body, the rule allows
only element reads at the loop index. Arrays and typed arrays qualify. The rule
skips known strings and plain objects and assumes that unknown inputs are arrays
or typed arrays.

Examples of **incorrect** code for this rule

```javascript
for (let index = 0; index < Math.min(names.length, scores.length); index++) {
  save(names[index], scores[index]);
}
```

Examples of **correct** code for this rule

```javascript
for (const [name, score] of Iterator.zip([names, scores])) {
  save(name, score);
}
```

## Fixes and suggestions

The editor suggestion uses `Iterator.zip()`, which requires runtime support.
Zipping initializes all iterators before entering the loop and reads each row
before the body runs. This can change when values are read if inputs have side
effects, use custom iterators, or are modified through aliases. Rslint reports
the loop without a suggestion if the replacement would remove a comment.

## Options

This rule has no options.

## Differences from upstream

Rslint reports the loop without a suggestion when a local `Math` or `Iterator`
would change what the replacement does. In this example, the proposed
`Iterator.zip()` call would use the `Iterator` parameter.

```javascript
function pair(Iterator, names, scores) {
  for (let index = 0; index < Math.min(names.length, scores.length); index++) {
    save(names[index], scores[index]);
  }
}
```

Rslint also withholds the suggestion when `languageOptions.globals` disables
`Iterator` or the loop's `Math`, or when the loop or its element reads are
inside a `with` statement.

Rslint ignores a bound such as `Math.min(index.length, names.length)` when
`index` is the loop counter. The counter starts at `0` and is a number.

Some suggested variable names differ. Rslint names the element from `cookies`
`cookie`, while Unicorn v79.0.0 names it `cooky`.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-iterator-zip.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-iterator-zip.js)
