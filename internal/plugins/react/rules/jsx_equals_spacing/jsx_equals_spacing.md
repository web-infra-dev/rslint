# react/jsx-equals-spacing

## Rule Details

Enforce or disallow spaces around the `=` sign in JSX attributes.

Examples of **incorrect** code with the default `"never"` option:

```jsx
<Foo name = "value" />
<Foo name= "value" />
<Foo name ="value" />
```

Examples of **correct** code with the default `"never"` option:

```jsx
<Foo name="value" />
```

## Options

- `"never"` (default): Disallow spaces around `=`.
- `"always"`: Require one space on each side of `=`.

## Original Documentation

- [react/jsx-equals-spacing](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-equals-spacing.md)
