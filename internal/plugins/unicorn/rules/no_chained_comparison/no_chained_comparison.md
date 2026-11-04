# no-chained-comparison

## Rule Details

Disallow chained comparisons such as `a < b < c`. JavaScript evaluates this as
`(a < b) < c`: the first comparison produces a boolean, which is then compared
with `c`. Use `&&` to compare each pair of operands.

Examples of **incorrect** code:

```javascript
if (a < b < c) {}
if (a === b === c) {}
```

Examples of **correct** code:

```javascript
if (a < b && b < c) {}
if (a === b && b === c) {}

// Intentional comparisons of boolean results are allowed.
if ((a > 0) === (b > 0)) {}
if ((a < b) === true) {}
if ((a > 0) !== Boolean(b)) {}
```

## Suggestions

The rule offers an editor suggestion to split simple chains with `&&`. This
changes the expression's behavior, so it is never applied as an automatic fix.

Suggestions are withheld when the middle operand has side effects, comments
would be duplicated or removed, or the replacement would change grouping.
Mixed ordering and equality operators, longer unparenthesized chains, and
ordering comparisons with a boolean literal also do not receive a suggestion.

## Options

This rule has no options.

## Differences from upstream

For the following code, rslint reports the chained comparison without offering
an editor suggestion. Unicorn v79.0.0 suggests a rewrite that can call `tag`
twice and change its side effects:

```javascript
a < tag`value` < c;
```

Evaluate the tagged template once, then write the intended comparisons:

```javascript
const middle = tag`value`;
a < middle && middle < c;
```

rslint distinguishes local classes from browser built-ins with the same name.
It reports the following comparison because the local `Headers.has()` returns
a number. Unicorn v79.0.0 can accept it with TypeScript type information or a
JSDoc annotation based on the class name:

```javascript
export class Headers {
  has() {
    return 0;
  }
}

(a < b) === new Headers().has('x');
```

## Original Documentation

- [eslint-plugin-unicorn: no-chained-comparison](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-chained-comparison.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-chained-comparison.js)
