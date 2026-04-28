# jsx-wrap-multilines

## Rule Details

Prevent missing parentheses around multiline JSX. Wrapping multiline JSX in parentheses improves readability and avoids potential issues with automatic semicolon insertion.

Examples of **incorrect** code for this rule:

```jsx
function App() {
  return <div>
    <span />
  </div>;
}
```

Examples of **correct** code for this rule:

```jsx
function App() {
  return (
    <div>
      <span />
    </div>
  );
}
```

## Options

An object with the following properties, each accepting:

- `"parens"`: Require parentheses around multiline JSX.
- `"parens-new-line"`: Require parentheses with opening `(` and closing `)` on separate lines from the JSX.
- `"never"`: Disallow parentheses around multiline JSX.
- `"ignore"`: Do not check this context.

Properties:

- `declaration` (default: `"parens"`): Variable declarations
- `assignment` (default: `"parens"`): Assignments
- `return` (default: `"parens"`): Return statements
- `arrow` (default: `"parens"`): Arrow function bodies
- `condition` (default: `"ignore"`): Ternary conditions
- `logical` (default: `"ignore"`): Logical expressions
- `prop` (default: `"ignore"`): JSX prop values

## Original Documentation

- [react/jsx-wrap-multilines](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-wrap-multilines.md)
