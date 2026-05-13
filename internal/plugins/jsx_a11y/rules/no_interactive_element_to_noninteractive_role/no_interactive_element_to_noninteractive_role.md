# no-interactive-element-to-noninteractive-role

## Rule Details

Inherently interactive HTML elements such as `<a href>`, `<button>`,
`<input>`, `<select>`, and `<textarea>` carry assistive-technology
semantics that non-interactive ARIA roles (`img`, `listitem`, `article`,
`presentation`, `none`, …) explicitly remove. Assigning a non-interactive
or presentation role to an interactive control hides its affordance from
screen-reader and keyboard users — they can no longer tell that the
element responds to interaction.

If a non-interactive role is genuinely desired for the surrounding
container, wrap the interactive element in a separate non-interactive
container instead of demoting the control itself.

The rule fires on every `role` JSX attribute when **all** of the following
hold:

- The resolved element name is in the HTML DOM set (custom React
  components are skipped — the rule does not know what low-level element
  they render).
- The attribute name is literally `role` (case-sensitive); namespaced
  attributes such as `mynamespace:role` are not checked.
- The element / `role` combination does not match an entry in the
  per-element allow-list (see Rule Options below).
- The element is inherently interactive (e.g. `<button>`, `<a href>`,
  `<input>`, `<select>`, `<textarea>`).
- The `role` attribute, when statically a literal string, resolves to a
  non-interactive role (`img`, `listitem`, `article`, …) OR to
  `presentation` / `none`.

Examples of **incorrect** code for this rule:

```jsx
<a href="http://example.com" role="img" />
<input type="text" role="listitem" />
<button role="article">Save</button>
```

Examples of **correct** code for this rule:

```jsx
<a href="http://example.com" role="button" />
<button>Save</button>
<div role="article">
  <button>Save</button>
</div>
```

## Rule Options

### Per-element allow-list

Type: `{ [tagName: string]: string[] }`. Default: not set. The upstream
`recommended` preset sets `tr: ["none", "presentation"]` and `canvas:
["img"]`; the `strict` preset omits the allow-list entirely.

Each key is an HTML element name and each value is an array of role
strings exempt from the rule for that element. A non-string entry in the
array is silently ignored. Non-array values are also ignored — only
`string[]` allow-lists are honored.

Examples of **correct** code with `{ "tr": ["none", "presentation"] }`:

```json
{ "jsx-a11y/no-interactive-element-to-noninteractive-role": ["error", { "tr": ["none", "presentation"] }] }
```

```jsx
<tr role="presentation" />
<tr role="none" />
```

Examples of **incorrect** code with `{ "tr": ["none", "presentation"] }`:

```json
{ "jsx-a11y/no-interactive-element-to-noninteractive-role": ["error", { "tr": ["none", "presentation"] }] }
```

```jsx
<tr role="img" />
```

## Resources

- [WCAG 4.1.2 — Name, Role, Value](https://www.w3.org/WAI/WCAG21/Understanding/name-role-value)
- [WAI-ARIA — Widget Roles](https://www.w3.org/TR/wai-aria-1.2/#widget_roles)
- [WAI-ARIA — Non-interactive Content Roles](https://www.w3.org/TR/wai-aria-1.2/#document_structure_roles)
- [MDN — Using ARIA: Roles, states, and properties](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Attributes)

## Original Documentation

- [eslint-plugin-jsx-a11y/no-interactive-element-to-noninteractive-role](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/no-interactive-element-to-noninteractive-role.md)
