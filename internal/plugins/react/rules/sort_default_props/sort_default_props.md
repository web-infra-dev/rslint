# sort-default-props

## Rule Details

Require `defaultProps` declarations to be sorted alphabetically. This rule checks class fields and assignments to `defaultProps` or `getDefaultProps`, including objects referenced through a variable initializer.

Each spread starts a new group. Property keys are compared by their source spelling, including quotes and escapes, and comparison is case-sensitive by default.

Examples of **incorrect** code for this rule:

```jsx
Button.defaultProps = {
  size: 'medium',
  disabled: false,
};

class Button extends React.Component {
  static defaultProps = {
    size: 'medium',
    disabled: false,
  };
}
```

Examples of **correct** code for this rule:

```jsx
Button.defaultProps = {
  disabled: false,
  size: 'medium',
};

Button.defaultProps = {
  size: 'medium',
  ...sharedDefaults,
  disabled: false,
  label: 'Submit',
};
```

The rule does not inspect objects returned from a `getDefaultProps()` method or follow chains of variable aliases.

## Options

### `ignoreCase`

When `true`, compare keys without regard to case. Defaults to `false`.

Examples of **incorrect** code with this option:

```json
{ "react/sort-default-props": ["error", { "ignoreCase": true }] }
```

```jsx
Button.defaultProps = {
  Size: 'medium',
  disabled: false,
};
```

Examples of **correct** code with this option:

```json
{ "react/sort-default-props": ["error", { "ignoreCase": true }] }
```

```jsx
Button.defaultProps = {
  disabled: false,
  Size: 'medium',
};
```

## Autofix

This rule does not provide an autofix or suggestions. Reordering properties can change evaluation order and spread behavior.

## Differences from ESLint

- When resolving a defaults variable, rslint follows lexical scope. ESLint can instead report an unrelated object in a child or sibling scope with the same variable name.
- For `const { defaults } = { z: 0, a: 0 }`, rslint does not check the containing object as defaults. ESLint reports its properties when `defaults` is assigned to `defaultProps`.

## Original Documentation

- [eslint-plugin-react: sort-default-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/sort-default-props.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/sort-default-props.js)
