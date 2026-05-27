# jsx-indent-props

Enforce props indentation in JSX.

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

## Options

The first option accepts:

- `"tab"` — tab-based indentation.
- a non-negative integer — N-space indentation.
- `"first"` — align each prop with the column of the first prop.
- an object with these keys:
  - `indentMode` — same value space as the positional option above (`"tab"`, `"first"`, or an integer).
  - `ignoreTernaryOperator` — boolean (default `false`). When `true`, props inside a JSX element nested in a `?:` consequent / alternate are NOT given an extra `indentMode` bump.

Examples of **incorrect** code for this rule with `["error", 2]`:

```json
{ "@stylistic/jsx-indent-props": ["error", 2] }
```

```jsx
<Hello
    firstName="John"
/>
```

Examples of **correct** code for this rule with `["error", 2]`:

```json
{ "@stylistic/jsx-indent-props": ["error", 2] }
```

```jsx
<Hello
  firstName="John"
/>
```

Examples of **correct** code for this rule with `["error", "tab"]`:

```json
{ "@stylistic/jsx-indent-props": ["error", "tab"] }
```

```jsx
<Hello
	firstName="John"
/>
```

Examples of **correct** code for this rule with `["error", 0]`:

```json
{ "@stylistic/jsx-indent-props": ["error", 0] }
```

```jsx
<Hello
firstName="John"
/>
```

Examples of **incorrect** code for this rule with `["error", "first"]`:

```json
{ "@stylistic/jsx-indent-props": ["error", "first"] }
```

```jsx
<Hello firstName="John"
  lastName="Doe"
/>
```

Examples of **correct** code for this rule with `["error", "first"]`:

```json
{ "@stylistic/jsx-indent-props": ["error", "first"] }
```

```jsx
<Hello firstName="John"
       lastName="Doe"
/>
```

Examples of **correct** code for this rule with `["error", { "indentMode": 2, "ignoreTernaryOperator": false }]` (default — props inside a `?:` branch receive an extra `indentMode` bump):

```json
{ "@stylistic/jsx-indent-props": ["error", { "indentMode": 2, "ignoreTernaryOperator": false }] }
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
{ "@stylistic/jsx-indent-props": ["error", { "indentMode": 2, "ignoreTernaryOperator": true }] }
```

```jsx
{condition
  ? <Hello
    firstName="John"
  />
  : null}
```

## Original Documentation

- [@stylistic/jsx-indent-props](https://eslint.style/rules/jsx-indent-props)
