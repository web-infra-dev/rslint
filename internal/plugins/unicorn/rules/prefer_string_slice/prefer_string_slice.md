# prefer-string-slice

## Rule Details

Prefer `String#slice()` over `String#substr()` and `String#substring()`.

`substr()` is deprecated. `substring()` clamps negative indices to zero and
swaps its arguments when the start is greater than the end. `slice()` supports
negative indices and uses the same indexing convention as `Array#slice()`.

Examples of **incorrect** code for this rule:

```javascript
const prefix = text.substr(0, 5);
const middle = text.substring(2, 5);
```

Examples of **correct** code for this rule:

```javascript
const prefix = text.slice(0, 5);
const middle = text.slice(2, 5);
const suffix = text.slice(-5);
```

The rule checks optional calls and optional member access. Computed method
names and receivers known to be non-strings are ignored.

Supported single-character `substring()` patterns, such as
`text.substring(index, index + 1)`, are left unchanged, matching upstream's
separation between string slicing and single-character access rules.

## Automatic fixes

The rule adjusts indices when it can convert the call automatically:

```javascript
text.substr(2, 3); // → text.slice(2, 5)
text.substring(5, 2); // → text.slice(2, 5)
text.substring(start); // → text.slice(Math.max(0, start))
```

An `undefined` end retains its meaning of “to the end”; a `null` end still
means zero. Fixes assume `.length` properties are non-negative. Calls whose
arguments cannot be converted, or whose conversion would discard comments
or repeat side effects, are reported without an automatic fix.

## Options

This rule has no options.

## Differences from upstream

rslint also honors explicit non-string JSDoc receiver types, including array
annotations that upstream may ignore without a TypeScript project.

rslint reports calls without an automatic fix when converting them could read a
getter or perform a coercion twice. For example, `text.substring(0, options.end)`
and `text.substr(0, +value)` are left for manual review. Store the end in a local
variable first to make an automatic fix available:

```javascript
const end = options.end;
text.substring(0, end);
```

Fixes that would introduce a locally shadowed or disabled global such as `Math`,
`undefined`, or `Infinity` are also withheld. Rename the conflicting binding to
allow the fix. Numeric rewrites that do not need those globals remain available.

## Original Documentation

- [eslint-plugin-unicorn documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-string-slice.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-string-slice.js)
