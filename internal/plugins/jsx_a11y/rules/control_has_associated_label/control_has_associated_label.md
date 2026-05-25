# control-has-associated-label

## Rule Details

Enforce that a control (an interactive element) has a text label.

There are several ways to supply a control with a text label:

- Provide text content inside the element.
- Set `aria-label` on the element.
- Set `aria-labelledby` and point it at an element with an accessible label.
- For an `img` inside the control, set its `alt` attribute.

The rule is permissive about runtime-resolvable labels: when the label
comes from an expression (`{maybeLabel}`) the rule assumes it will
provide a label.

Examples of **incorrect** code for this rule:

```jsx
<button />
<button><span /></button>
<button><span title="This is not a real label" /></button>
<a href="#" />
<area href="#" />
<menuitem />
<option />
<th />
<div role="button" />
<div role="checkbox" />
```

Examples of **correct** code for this rule:

```jsx
<button>Save</button>
<button><span>Save</span></button>
<button><img alt="Save" /></button>
<button aria-label="Save" />
<button aria-labelledby="js_1" />
<button>{maybeLabel}</button>
<a href="#">Save</a>
<th>Save</th>
<div role="button">Save</div>
<div role="checkbox" aria-label="Save" />
<div role="link" aria-labelledby="js_1" />
```

## Rule Options

This rule takes one optional object argument:

```json
{
  "labelAttributes": ["label"],
  "controlComponents": ["CustomControl"],
  "ignoreElements": [
    "audio",
    "canvas",
    "embed",
    "input",
    "textarea",
    "tr",
    "video"
  ],
  "ignoreRoles": [
    "grid",
    "listbox",
    "menu",
    "menubar",
    "radiogroup",
    "row",
    "tablist",
    "toolbar",
    "tree",
    "treegrid"
  ],
  "depth": 3
}
```

- `labelAttributes` — extra attribute names to treat as labels in
  addition to `alt`, `aria-label`, and `aria-labelledby`. Useful when a
  custom component renders a label from a specific prop (e.g. `label`).
- `controlComponents` — custom React component names to treat as
  interactive controls. The top-level trigger uses exact matching; the
  recursive children-walk's React-component fallback uses minimatch
  glob patterns (mirrors upstream).
- `ignoreElements` — elements that should not be considered interactive
  controls (e.g. `input`, `textarea`, where labels are commonly supplied
  by a wrapping `<label>`).
- `ignoreRoles` — ARIA roles that should not be considered interactive
  controls.
- `depth` (default `2`, max `25`) — how deep within the JSX subtree the
  rule should look for an accessible label.

The `link` tag is **always** ignored — the upstream rule hard-codes this
exemption and it cannot be disabled.

Examples of **incorrect** code for this rule with
`{ "controlComponents": ["CustomControl"] }`:

```json
{
  "jsx-a11y/control-has-associated-label": [
    "error",
    { "controlComponents": ["CustomControl"] }
  ]
}
```

```jsx
<CustomControl />
<CustomControl><span /></CustomControl>
```

Examples of **correct** code for this rule with
`{ "depth": 3, "controlComponents": ["CustomControl"] }`:

```json
{
  "jsx-a11y/control-has-associated-label": [
    "error",
    { "depth": 3, "controlComponents": ["CustomControl"] }
  ]
}
```

```jsx
<CustomControl><span><span>Save</span></span></CustomControl>
<CustomControl aria-label="Save" />
```

Examples of **correct** code for this rule with
`{ "labelAttributes": ["label"] }`:

```json
{
  "jsx-a11y/control-has-associated-label": [
    "error",
    { "labelAttributes": ["label"] }
  ]
}
```

```jsx
<button><span label="Save" /></button>
```

## Accessibility guidelines

- [WCAG 1.3.1 — Info and Relationships](https://www.w3.org/WAI/WCAG21/Understanding/info-and-relationships)
- [WCAG 3.3.2 — Labels or Instructions](https://www.w3.org/WAI/WCAG21/Understanding/labels-or-instructions)
- [WCAG 4.1.2 — Name, Role, Value](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)

### Resources

- [MDN — `aria-label`](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Attributes/aria-label)
- [MDN — `aria-labelledby`](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Attributes/aria-labelledby)
- [WAI-ARIA 1.2 — Accessible Name and Description Computation](https://www.w3.org/TR/wai-aria-1.2/#namecalculation)

## Original Documentation

- [eslint-plugin-jsx-a11y/control-has-associated-label](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/control-has-associated-label.md)
