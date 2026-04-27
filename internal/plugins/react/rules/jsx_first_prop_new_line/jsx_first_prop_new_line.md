# jsx-first-prop-new-line

## Rule Details

Enforce the position of the first prop in a JSX element. Useful for enforcing consistent formatting of JSX props.

Examples of **incorrect** code with `"multiline-multiprop"` (default):

```jsx
<Hello foo="bar"
  baz="quux"
/>
```

Examples of **correct** code with `"multiline-multiprop"` (default):

```jsx
<Hello foo="bar" baz="quux" />
<Hello
  foo="bar"
  baz="quux"
/>
```

## Options

- `"always"`: First prop must always be on a new line.
- `"never"`: First prop must never be on a new line.
- `"multiline"`: First prop must be on a new line when the JSX tag spans multiple lines.
- `"multiline-multiprop"` (default): First prop must be on a new line when the JSX tag spans multiple lines and has multiple props.
- `"multiprop"`: First prop must be on a new line when the JSX tag has multiple props.

## Original Documentation

- [react/jsx-first-prop-new-line](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-first-prop-new-line.md)
