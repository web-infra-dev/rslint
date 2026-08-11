# jsx-equals-spacing

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

- [eslint-plugin-react: jsx-equals-spacing](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-equals-spacing.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-equals-spacing.js)
