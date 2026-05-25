# jsx-indent-props

## Rule Details

This rule enforces a consistent indentation style for JSX props. The default style is `4 spaces`.

Examples of **incorrect** code for this rule:

```jsx
<Hello
  firstName="John"
/>
```

Examples of **correct** code for this rule:

```jsx
<Hello
    firstName="John"
/>
```

## Rule Options

The first option accepts:

- `"tab"` — tab-based indentation.
- a non-negative integer — N-space indentation.
- `"first"` — align each prop with the column of the first prop.
- an object with these keys:
  - `indentMode` — same value space as the positional option above (`"tab"`, `"first"`, or an integer).
  - `ignoreTernaryOperator` — boolean (default `false`). When `true`, props inside a JSX element nested in a `?:` consequent / alternate are NOT given an extra `indentMode` bump.

Examples of **incorrect** code for this rule with `["error", 2]`:

```json
{ "react/jsx-indent-props": ["error", 2] }
```

```jsx
<Hello
    firstName="John"
/>
```

Examples of **correct** code for this rule with `["error", 2]`:

```json
{ "react/jsx-indent-props": ["error", 2] }
```

```jsx
<Hello
  firstName="John"
/>
```

Examples of **correct** code for this rule with `["error", "tab"]`:

```json
{ "react/jsx-indent-props": ["error", "tab"] }
```

```jsx
<Hello
	firstName="John"
/>
```

Examples of **correct** code for this rule with `["error", 0]`:

```json
{ "react/jsx-indent-props": ["error", 0] }
```

```jsx
<Hello
firstName="John"
/>
```

Examples of **incorrect** code for this rule with `["error", "first"]`:

```json
{ "react/jsx-indent-props": ["error", "first"] }
```

```jsx
<Hello firstName="John"
  lastName="Doe"
/>
```

Examples of **correct** code for this rule with `["error", "first"]`:

```json
{ "react/jsx-indent-props": ["error", "first"] }
```

```jsx
<Hello firstName="John"
       lastName="Doe"
/>
```

Examples of **correct** code for this rule with `["error", { "indentMode": 2, "ignoreTernaryOperator": false }]` (default — props inside a `?:` branch receive an extra `indentMode` bump):

```json
{ "react/jsx-indent-props": ["error", { "indentMode": 2, "ignoreTernaryOperator": false }] }
```

```jsx
{condition
  ? <Hello
      firstName="John"
    />
  : null}
```

Examples of **correct** code for this rule with `["error", { "indentMode": 2, "ignoreTernaryOperator": true }]` (no extra bump inside `?:`):

```json
{ "react/jsx-indent-props": ["error", { "indentMode": 2, "ignoreTernaryOperator": true }] }
```

```jsx
{condition
  ? <Hello
    firstName="John"
  />
  : null}
```

## When Not To Use It

If you are not using JSX, you can disable this rule. If you use a code formatter (e.g. Prettier) that already enforces JSX prop indentation, prefer relying on the formatter instead.

## Original Documentation

- [react/jsx-indent-props](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-indent-props.md)
