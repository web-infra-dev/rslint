# no-array-sort

## Rule Details

Prefer `Array#toSorted()` over the mutating `Array#sort()` method. The rule provides editor suggestions, not an automatic fix, because changing whether the original array is mutated requires an explicit choice.

Examples of **incorrect** code:

```javascript
const result = array.sort();
const copy = [...array].sort();
```

Examples of **correct** code:

```javascript
const result = array.toSorted();
const copy = [...iterable].toSorted();
```

For a single spread, suggestions let you either remove the spread when its value is an array or retain it for another iterable. Optional member access is preserved; optional calls and computed method names are not checked.

## Options

### `allowExpressionStatement`

Type: `boolean`

Default: `true`

By default, a direct `array.sort()` expression statement is allowed. A spread-copy mutation is still reported. Set `allowExpressionStatement` to `false` to report direct expression statements too.

```javascript
{
  'unicorn/no-array-sort': ['error', { allowExpressionStatement: false }]
}
```

Calls with a clearly non-function comparator argument are not treated as array sorting.

## Original Documentation

- [eslint-plugin-unicorn: no-array-sort](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-array-sort.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-array-sort.js)
