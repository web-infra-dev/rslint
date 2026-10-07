# no-new-array

## Rule Details

Disallow `new Array()` with one argument because it is unclear whether the argument is an array length or a single element.

The rule complements `no-array-constructor`, which handles constructors with zero or multiple arguments.

Examples of **incorrect** code:

```javascript
const array = new Array(onlyElement);
const items = new Array(...values);
```

Examples of **correct** code:

```javascript
const array = [onlyElement];
const items = [...values];
```

Known non-numeric arguments can be automatically replaced with an array literal. A spread argument receives an editor suggestion. Comments inside the constructor prevent both kinds of edits.

Numeric and unknown arguments are reported without edits. `new Array(3)` creates three holes, while `Array.from({ length: 3 })` creates three `undefined` elements; these behave differently with methods such as `filter()`.

## Options

This rule has no options. It provides automatic fixes and editor suggestions and is enabled at `error` severity in `unicornPlugin.configs.recommended`.

## Differences from upstream

When `Array` refers to a local declaration, parameter, or import, or is disabled through `languageOptions.globals`, rslint reports the diagnostic without a fix or suggestion. This preserves custom constructor behavior, such as setting `result.value` in `new Array('x')`.

Constructors with explicit TypeScript type arguments also receive no fix or suggestion. For example, changing `new Array<string | null>('x')` to `['x']` would infer `string[]` and reject a later `push(null)`. To use a literal while retaining the element type, write `const values: (string | null)[] = ['x'];`.

For built-in objects, functions, and symbols, such as `new Array(Math)`, `new Array(Array)`, or `new Array(Symbol.iterator)`, rslint reports the same diagnostic but does not offer upstream's automatic fix. This also applies to their `typeof` expressions, such as `new Array(typeof Array)`.

Replace these constructors with array literals manually, for example `[Symbol.iterator]` or `[typeof Array]`, to make the intended single element explicit.

## Original Documentation

- [Upstream rule documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/no-new-array.md)
- [Upstream rule source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-new-array.js)
