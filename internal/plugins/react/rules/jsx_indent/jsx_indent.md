# jsx-indent

## Rule Details

This rule enforces a consistent indentation style for JSX. The default style is `4 spaces`.

Examples of **incorrect** code for this rule:

```jsx
<App>
  <Hello />
</App>
```

Examples of **correct** code for this rule:

```jsx
<App>
    <Hello />
</App>
```

## Rule Options

The first option is either the string `"tab"` for tab-based indentation, or a non-negative integer for space-based indentation.

The second option is an object with two boolean keys (each defaulting to `false`):

- `checkAttributes` — also enforce indentation inside JSX-expression attribute values.
- `indentLogicalExpressions` — also indent JSX nested inside the right side of a logical (`&&`, `||`, `??`) expression.

Examples of **incorrect** code for this rule with `["error", 2]`:

```json
{ "react/jsx-indent": ["error", 2] }
```

```jsx
<App>
    <Hello />
</App>
```

Examples of **correct** code for this rule with `["error", 2]`:

```json
{ "react/jsx-indent": ["error", 2] }
```

```jsx
<App>
  <Hello />
</App>
```

Examples of **correct** code for this rule with `["error", "tab"]`:

```json
{ "react/jsx-indent": ["error", "tab"] }
```

```jsx
<App>
	<Hello />
</App>
```

Examples of **correct** code for this rule with `["error", 0]`:

```json
{ "react/jsx-indent": ["error", 0] }
```

```jsx
<App>
<Hello />
</App>
```

Examples of **incorrect** code for this rule with `["error", 2, { "indentLogicalExpressions": true }]`:

```json
{ "react/jsx-indent": ["error", 2, { "indentLogicalExpressions": true }] }
```

```jsx
<App>
  {condition && (
  <Hello />
  )}
</App>
```

Examples of **correct** code for this rule with `["error", 2, { "indentLogicalExpressions": true }]`:

```json
{ "react/jsx-indent": ["error", 2, { "indentLogicalExpressions": true }] }
```

```jsx
<App>
  {condition && (
    <Hello />
  )}
</App>
```

## When Not To Use It

If you are not using JSX, you can disable this rule. If you use a code formatter (e.g. Prettier) that already enforces JSX indentation, prefer relying on the formatter instead.

## Original Documentation

- [react/jsx-indent](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-indent.md)
