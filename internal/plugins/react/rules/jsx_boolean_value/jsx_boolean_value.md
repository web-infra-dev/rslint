# react/jsx-boolean-value

## Rule Details

Enforce boolean attributes notation in JSX. When a prop is `true`, it can be written as either `<Foo bar />` or `<Foo bar={true} />`.

Examples of **incorrect** code with the default `"never"` option:

```jsx
<Foo bar={true} />
```

Examples of **correct** code with the default `"never"` option:

```jsx
<Foo bar />
```

## Options

- First argument: `"never"` (default) or `"always"`
  - `"never"`: Enforce shorthand (`<Foo bar />`)
  - `"always"`: Enforce explicit value (`<Foo bar={true} />`)
- Second argument (optional): An object with:
  - `"never"`: Array of prop names that should use shorthand. Only valid when first argument is `"always"`.
  - `"always"`: Array of prop names that should use explicit `={true}`. Only valid when first argument is `"never"`.
  - `"assumeUndefinedIsFalse"`: When `true`, `={false}` props are reported for removal (omitting a boolean prop is treated as `false`).

## Original Documentation

- [react/jsx-boolean-value](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-boolean-value.md)
