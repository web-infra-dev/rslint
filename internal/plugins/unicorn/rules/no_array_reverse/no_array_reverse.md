# no-array-reverse

## Rule Details

Prefer `Array#toReversed()` over the mutating `Array#reverse()` method. The rule provides editor suggestions, not an automatic fix, because changing whether the original array is mutated requires an explicit choice.

Examples of **incorrect** code:

```javascript
const result = array.reverse();
const copy = [...array].reverse();
```

Examples of **correct** code:

```javascript
const result = array.toReversed();
const copy = [...iterable].toReversed();
```

For a single spread, suggestions let you either remove the spread when its value is an array or retain it for another iterable. Optional member access is preserved; optional calls and computed method names are not checked.

## Options

### `allowExpressionStatement`

Type: `boolean`

Default: `true`

By default, a direct `array.reverse()` expression statement is allowed. A spread-copy mutation is still reported. Set `allowExpressionStatement` to `false` to report direct expression statements too.

```javascript
{
  'unicorn/no-array-reverse': ['error', { allowExpressionStatement: false }]
}
```

## Original Documentation

- [eslint-plugin-unicorn: no-array-reverse](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-array-reverse.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-array-reverse.js)
