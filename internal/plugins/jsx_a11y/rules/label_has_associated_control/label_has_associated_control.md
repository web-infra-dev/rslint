# label-has-associated-control

## Rule Details

Enforce that a `<label>` tag has a text label AND an associated form control.

There are two supported ways to associate a label with a control:

- **Nesting** — wrap the control in the label.
- **`htmlFor`** — give the label an `htmlFor` attribute pointing at the
  control's DOM id.

This rule checks that any `<label>` tag (or a configured custom label
component) either (1) wraps a form control or (2) declares an `htmlFor`
attribute, AND that the label has accessible text content.

Examples of **incorrect** code for this rule:

```jsx
<label htmlFor="js_id" />
<label htmlFor="js_id"><input /></label>
<label htmlFor="js_id"><textarea /></label>
<label></label>
<label>A label</label>
<div><label /><input /></div>
<div><label>A label</label><input /></div>
```

Examples of **correct** code for this rule:

```jsx
<label htmlFor="js_id">A label</label>
<label htmlFor="js_id" aria-label="A label" />
<label htmlFor="js_id" aria-labelledby="A label" />
<label>A label<input /></label>
<label>A label<textarea /></label>
<label><img alt="A label" /><input /></label>
<label htmlFor="selectInput">Some text<select id="selectInput" /></label>
```

## Rule Options

This rule takes one optional object argument:

```json
{
  "jsx-a11y/label-has-associated-control": [
    "error",
    {
      "labelComponents": ["CustomLabel"],
      "labelAttributes": ["label"],
      "controlComponents": ["CustomInput"],
      "assert": "either",
      "depth": 2
    }
  ]
}
```

- `labelComponents` (default `[]`) — custom React component names treated
  as a `<label>`. Glob patterns are supported via minimatch (e.g. `*Label`
  matches `MUILabel`, `????Label` matches `LinkLabel` but not `CustomLabel`).
- `labelAttributes` (default `[]`) — extra attribute names that count as
  accessible text on the label (or any descendant within `depth`), in
  addition to `alt` / `aria-label` / `aria-labelledby`. Useful when a
  custom component renders its label from a string prop.
- `controlComponents` (default `[]`) — custom React component names that
  count as a nested form control. These are added to the built-in list:
  `input`, `meter`, `output`, `progress`, `select`, `textarea`. Glob
  patterns are supported (e.g. `Custom*` matches `CustomInput`).
- `assert` (default `"either"`) — which association strategy to require.
  One of `"htmlFor"`, `"nesting"`, `"both"`, `"either"`:
  - `"htmlFor"` — the label must have a valid `htmlFor` attribute.
  - `"nesting"` — the label must wrap a form control.
  - `"both"` — the label must satisfy both conditions.
  - `"either"` — the label must satisfy at least one condition.
- `depth` (default `2`, max `25`) — how deep within a `<label>` (or label
  component) the rule will recurse to find accessible text or a nested
  form control. Capped at 25 to prevent pathological traversal.

Examples of **correct** code for this rule with
`{ "labelComponents": ["CustomLabel"] }`:

```json
{
  "jsx-a11y/label-has-associated-control": [
    "error",
    { "labelComponents": ["CustomLabel"] }
  ]
}
```

```jsx
<CustomLabel htmlFor="js_id" aria-label="A label" />
```

Examples of **correct** code for this rule with
`{ "labelAttributes": ["label"] }`:

```json
{
  "jsx-a11y/label-has-associated-control": [
    "error",
    { "labelAttributes": ["label"] }
  ]
}
```

```jsx
<label htmlFor="js_id" label="A label" />
```

Examples of **correct** code for this rule with
`{ "controlComponents": ["CustomInput"] }`:

```json
{
  "jsx-a11y/label-has-associated-control": [
    "error",
    { "controlComponents": ["CustomInput"] }
  ]
}
```

```jsx
<label>A label<CustomInput /></label>
```

Examples of **incorrect** code for this rule with `{ "assert": "both" }`:

```json
{
  "jsx-a11y/label-has-associated-control": ["error", { "assert": "both" }]
}
```

```jsx
<label htmlFor="js_id">A label</label>
<label>A label<input /></label>
```

Examples of **correct** code for this rule with `{ "assert": "both" }`:

```json
{
  "jsx-a11y/label-has-associated-control": ["error", { "assert": "both" }]
}
```

```jsx
<label htmlFor="js_id" aria-label="A label"><input /></label>
```

Examples of **correct** code for this rule with `{ "depth": 3 }`:

```json
{
  "jsx-a11y/label-has-associated-control": ["error", { "depth": 3 }]
}
```

```jsx
<label><span><span>A label<input /></span></span></label>
```

## Settings

The rule reads two `settings['jsx-a11y']` keys:

- `attributes.for` — list of attribute names recognized as equivalent to
  `htmlFor`. Defaults to `["htmlFor"]`. When set, the user list
  **replaces** the default — include `"htmlFor"` explicitly to keep it
  active.
- `components` — a `{ ComponentName: "html-element" }` map. With
  `{ "CustomLabel": "label" }`, `<CustomLabel>` is treated as `<label>`
  for the listener gate.

```json
{
  "settings": {
    "jsx-a11y": {
      "attributes": { "for": ["htmlFor", "for"] },
      "components": { "CustomLabel": "label", "CustomInput": "input" }
    }
  }
}
```

```jsx
<label for="js_id" aria-label="A label" />
<CustomLabel htmlFor="js_id" aria-label="A label" />
```

These mirror the upstream `eslint-plugin-jsx-a11y` settings exactly.

## Accessibility guidelines

- [WCAG 1.3.1 — Info and Relationships](https://www.w3.org/WAI/WCAG21/Understanding/info-and-relationships)
- [WCAG 3.3.2 — Labels or Instructions](https://www.w3.org/WAI/WCAG21/Understanding/labels-or-instructions)
- [WCAG 4.1.2 — Name, Role, Value](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)

### Resources

- [MDN — `<label>`](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/label)
- [MDN — `aria-label`](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Attributes/aria-label)
- [MDN — `aria-labelledby`](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Attributes/aria-labelledby)
- [WAI-ARIA 1.2 — Accessible Name and Description Computation](https://www.w3.org/TR/wai-aria-1.2/#namecalculation)

## Original Documentation

- [eslint-plugin-jsx-a11y/label-has-associated-control](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/label-has-associated-control.md)
