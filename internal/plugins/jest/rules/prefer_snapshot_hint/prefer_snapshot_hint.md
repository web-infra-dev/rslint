# prefer-snapshot-hint

## Rule Details

Prefer including a descriptive hint with external snapshots. Hints make snapshots easier to identify during review and updates.

The rule checks `toMatchSnapshot` and `toThrowErrorMatchingSnapshot`. Inline snapshots are ignored. The default `"multi"` mode requires hints in groups containing multiple external snapshot assertions; `"always"` requires a hint on every external snapshot assertion. Function expressions, arrow functions, methods, and accessors establish groups that include nested helper expressions, while test and suite registrations reset the grouping boundary. Function declarations do not establish separate groups.

For `toMatchSnapshot`, a single string literal or template without interpolation counts as a hint, and two arguments count as properties plus a hint. A single variable or interpolated template is reported, regardless of its type. `toThrowErrorMatchingSnapshot` accepts a single hint argument. This rule does not require type information or provide fixes.

## Incorrect

```js
test('cart totals', () => {
  expect(cart.subtotal).toMatchSnapshot();
  expect(cart.total).toMatchSnapshot();
});
```

## Correct

```js
test('cart totals', () => {
  expect(cart.subtotal).toMatchSnapshot('before tax');
  expect(cart.total).toMatchSnapshot('after tax');
});
```

## Options

```json
{
  "jest/prefer-snapshot-hint": ["error", "always"]
}
```

## Original Documentation

- [eslint-plugin-jest: prefer-snapshot-hint](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/docs/rules/prefer-snapshot-hint.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/prefer-snapshot-hint.ts)
