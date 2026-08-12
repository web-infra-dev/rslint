# self-closing-comp

## Rule Details

Enforce components without children to be self-closing. This applies to both custom components and HTML elements.

Examples of **incorrect** code for this rule:

```jsx
<Hello name="John"></Hello>
<div className="content"></div>
```

Examples of **correct** code for this rule:

```jsx
<Hello name="John" />
<div className="content" />
<div>Children</div>
```

## Options

- `component` (default: `true`): Whether to enforce self-closing for custom components.
- `html` (default: `true`): Whether to enforce self-closing for HTML elements.

## Original Documentation

- [eslint-plugin-react: self-closing-comp](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/self-closing-comp.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/self-closing-comp.js)
