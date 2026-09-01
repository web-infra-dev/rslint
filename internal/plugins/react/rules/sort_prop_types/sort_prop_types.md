# sort-prop-types

## Rule Details

Enforce alphabetical ordering of React `propTypes` declarations. Options can
also require required props first, callback props last, and inspect
`PropTypes.shape` and TypeScript component prop types.

Examples of **incorrect** code for this rule:

```jsx
Component.propTypes = { zebra: PropTypes.string, alpha: PropTypes.string };
```

Examples of **correct** code for this rule:

```jsx
Component.propTypes = { alpha: PropTypes.string, zebra: PropTypes.string };
```

## Original Documentation

- [eslint-plugin-react: sort-prop-types](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/sort-prop-types.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/sort-prop-types.js)
