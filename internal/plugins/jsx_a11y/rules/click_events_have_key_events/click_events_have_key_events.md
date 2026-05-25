# click-events-have-key-events

## Rule Details

Enforce that a visible, non-interactive element with an `onClick` handler
also declares at least one keyboard event listener (`onKeyDown`,
`onKeyUp`, or `onKeyPress`). Pairing a pointer event with a keyboard
counterpart keeps the interaction reachable for keyboard-only users and
assistive technologies.

The rule reports a JSX opening element when **all** of the following hold:

- The resolved element name is in the HTML DOM set (custom React
  components are skipped — the rule does not know what low-level element
  they render).
- The element is not hidden from screen readers (no `aria-hidden={true}`
  / `aria-hidden="true"`, not an `<input type="hidden">`).
- The `role` attribute, when statically a literal string, is not
  `presentation` or `none`.
- The element is not inherently interactive (e.g. `<button>`, `<a href>`,
  `<select>`, `<input type="text">`).
- The element declares an `onClick` attribute (case-insensitively, value
  irrelevant — boolean form, `null`, `undefined` all count as present).
- The element declares none of `onKeyDown`, `onKeyUp`, `onKeyPress`
  (case-insensitively, as a direct attribute — spread attributes are
  opaque).

This rule takes no arguments.

Examples of **incorrect** code for this rule:

```jsx
<div onClick={() => {}} />
<section onClick={() => {}} />
<a onClick={() => {}} />
<div onClick={() => {}} {...props} />
```

Examples of **correct** code for this rule:

```jsx
<div onClick={() => {}} onKeyDown={handleKeyDown} />
<div onClick={() => {}} onKeyUp={handleKeyUp} />
<div onClick={() => {}} onKeyPress={handleKeyPress} />
<button onClick={() => {}} />
<a onClick={() => {}} href="/profile" />
<div onClick={() => {}} aria-hidden="true" />
<div onClick={() => {}} role="presentation" />
<MyComponent onClick={() => {}} />
```

## Resources

- [WCAG 2.1.1 — Keyboard](https://www.w3.org/WAI/WCAG21/Understanding/keyboard)
- [WAI-ARIA — `aria-hidden`](https://www.w3.org/TR/wai-aria-1.2/#aria-hidden)
- [MDN — Keyboard-navigable JavaScript widgets](https://developer.mozilla.org/en-US/docs/Web/Accessibility/Keyboard-navigable_JavaScript_widgets)

## Original Documentation

- [eslint-plugin-jsx-a11y/click-events-have-key-events](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/click-events-have-key-events.md)
