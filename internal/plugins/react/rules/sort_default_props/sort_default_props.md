# sort-default-props

## Rule Details

Require `defaultProps` declarations to be sorted alphabetically. This rule checks class fields and direct or logical assignments (`=`, `||=`, `&&=`, and `??=`) to `defaultProps` or `getDefaultProps`, including objects referenced through a variable initializer.

Property names can use dot access, quoted names, or computed string and template literals, such as `Button.defaultProps`, `Button["defaultProps"]`, or `static ["defaultProps"]`. Dynamic property names and private fields are ignored.

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

Defaults variables are resolved in their lexical value scope. Type declarations do not hide runtime variables. Exported defaults remain visible across declarations of the same namespace. The first value definition is used when a name has several declarations; later assignments and chains of aliases are not followed.

The rule does not inspect objects returned from a `getDefaultProps()` method. Reading or comparing `defaultProps`, or using it as a loop target, does not declare defaults.

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

- Defaults variables resolve in lexical value scope, including computed method names, parameter defaults, and namespace exports. ESLint can select an unrelated child-scope declaration.
- A same-name type or interface does not suppress a variable's defaults check. An earlier `var` definition remains visible when a function has the same name.
- For `const { defaults } = { z: 0, a: 0 }`, rslint does not check the containing object as defaults. ESLint can report its properties instead.
- Quoted and computed literal declarations such as `Button["defaultProps"]` are checked. Dynamic identifiers named `defaultProps` and private fields are ignored.
- Reads, comparisons, arithmetic assignments, and `for-in`/`for-of` sources are not defaults declarations. ESLint can report unrelated objects in those expressions.

## Original Documentation

- [eslint-plugin-react: sort-default-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/sort-default-props.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/sort-default-props.js)
