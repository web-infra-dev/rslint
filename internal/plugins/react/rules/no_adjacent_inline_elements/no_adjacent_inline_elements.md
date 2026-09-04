# no-adjacent-inline-elements

## Rule Details

This rule disallows adjacent inline HTML elements without whitespace between
them. Without a separating space, inline elements can run into each other when
the content is viewed without styling.

Examples of **incorrect** code for this rule:

```jsx
<div><a></a><a></a></div>
<div><a></a><span></span></div>

React.createElement("div", undefined, [React.createElement("a"), React.createElement("span")]);
```

Examples of **correct** code for this rule:

```jsx
<div><div></div><div></div></div>
<div><a></a> <a></a></div>

React.createElement("div", undefined, [React.createElement("a"), " ", React.createElement("a")]);
```

## When Not To Use It

Disable this rule when adjacent inline elements are intentional or spacing is
provided by styling.

## Original Documentation

- [eslint-plugin-react: no-adjacent-inline-elements](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/no-adjacent-inline-elements.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/no-adjacent-inline-elements.js)
