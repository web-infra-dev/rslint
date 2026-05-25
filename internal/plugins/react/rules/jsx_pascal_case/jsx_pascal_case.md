# jsx-pascal-case

Enforce PascalCase for user-defined JSX components.

## Rule Details

JSX tag names that start with a lowercase letter are interpreted by React as
HTML intrinsics (`<div>`, `<span>`, …) and are skipped. For user components
(`<TestComponent>`), this rule requires PascalCase: the first character must
be an upper-case letter, the remaining characters must be alphanumeric, and
at least one lower-case letter or digit must follow.

Examples of **incorrect** code for this rule:

```jsx
<Test_component />
```

```jsx
<TEST_COMPONENT />
```

Examples of **correct** code for this rule:

```jsx
<div />
```

```jsx
<TestComponent />
```

```jsx
<TestComponent>
  <div />
</TestComponent>
```

```jsx
<CSSTransitionGroup />
```

## Rule Options

```json
{
  "react/jsx-pascal-case": [
    "error",
    {
      "allowAllCaps": false,
      "allowNamespace": false,
      "allowLeadingUnderscore": false,
      "ignore": []
    }
  ]
}
```

- `allowAllCaps` (default `false`): allow SCREAMING\_SNAKE\_CASE component names.
- `allowNamespace` (default `false`): skip the check for parts after the first
  dot or colon in a namespaced / member-access tag.
- `allowLeadingUnderscore` (default `false`): strip a single leading `_` before
  running the PascalCase / all-caps check.
- `ignore` (default `[]`): array of names to exempt. Entries are matched as
  minimatch-style globs — supports `*`, `?`, character classes, and extglob
  groups such as `+(a|b)`.

### `allowAllCaps`

Examples of **correct** code for this rule, when `allowAllCaps` is `true`:

```json
{ "react/jsx-pascal-case": ["error", { "allowAllCaps": true }] }
```

```jsx
<ALLOWED />
<TEST_COMPONENT />
```

### `allowNamespace`

Examples of **correct** code for this rule, when `allowNamespace` is `true`:

```json
{ "react/jsx-pascal-case": ["error", { "allowNamespace": true }] }
```

```jsx
<Allowed.div />
<TestComponent.p />
```

### `allowLeadingUnderscore`

Examples of **correct** code for this rule, when `allowLeadingUnderscore` is
`true`:

```json
{ "react/jsx-pascal-case": ["error", { "allowLeadingUnderscore": true }] }
```

```jsx
<_AllowedComponent />
<_AllowedComponent>
  <div />
</_AllowedComponent>
```

## Differences from ESLint

- `ignore` entries do not support the `!(a|b)` negative extglob syntax.

## Original Documentation

- ESLint-plugin-react rule: [https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-pascal-case.md](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/jsx-pascal-case.md)
- Source: [https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/jsx-pascal-case.js](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/jsx-pascal-case.js)
