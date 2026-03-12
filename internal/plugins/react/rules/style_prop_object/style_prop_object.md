# react/style-prop-object

## Rule Details

Enforce that the `style` prop value is an object. In React, the `style` prop expects a JavaScript object, not a CSS string.

Examples of **incorrect** code for this rule:

```jsx
<div style="color: red" />
<div style={"color: red"} />
```

Examples of **correct** code for this rule:

```jsx
<div style={{ color: 'red' }} />
<div style={styles} />
```

## Options

- `allow`: An array of component names that are allowed to use a non-object `style` prop.

## Limitations

- Only detects `React.createElement(...)` calls. Destructured `createElement` (e.g. `import { createElement } from 'react'`) and custom pragma (e.g. `Preact.h`) are not supported.

## Original Documentation

- [react/style-prop-object](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/style-prop-object.md)
