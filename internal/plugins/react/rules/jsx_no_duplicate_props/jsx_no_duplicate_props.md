# jsx-no-duplicate-props

Disallow duplicate properties in JSX.

## Rule Details

Creating JSX elements with duplicate props can cause unexpected behavior in your application.

Examples of **incorrect** code for this rule:

```jsx
<Hello name="John" name="John" />;
```

Examples of **correct** code for this rule:

```jsx
<Hello first="John" last="Doe" />;
```

## Rule Options

```json
{ "react/jsx-no-duplicate-props": ["error", { "ignoreCase": true }] }
```

### `ignoreCase`

When `true` the rule ignores the case of the props. Defaults to `false`.

Examples of **incorrect** code for this rule with `{ "ignoreCase": true }`:

```json
{ "react/jsx-no-duplicate-props": ["error", { "ignoreCase": true }] }
```

```jsx
<Hello name="John" Name="John" />;
```

## Original Documentation

- [eslint-plugin-react docs](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-no-duplicate-props.md)
