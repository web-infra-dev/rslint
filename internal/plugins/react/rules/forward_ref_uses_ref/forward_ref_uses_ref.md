# forward-ref-uses-ref

Require all `forwardRef` components to include a `ref` parameter.

## Rule Details

Components wrapped with `forwardRef` receive props as the first parameter and the forwarded ref as the second parameter. This rule reports `forwardRef` callbacks that only declare the props parameter.

Examples of **incorrect** code for this rule:

```jsx
const Button = forwardRef((props) => <button {...props} />);
```

```jsx
const Button = React.forwardRef(function Button(props) {
  return <button {...props} />;
});
```

Examples of **correct** code for this rule:

```jsx
const Button = forwardRef((props, ref) => <button ref={ref} {...props} />);
```

```jsx
function Button(props) {
  return <button {...props} />;
}
```

## Options

This rule has no options.

## Original Documentation

- [eslint-plugin-react: forward-ref-uses-ref](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/forward-ref-uses-ref.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/forward-ref-uses-ref.js)
