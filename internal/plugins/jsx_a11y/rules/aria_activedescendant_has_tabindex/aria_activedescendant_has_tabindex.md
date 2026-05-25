# aria-activedescendant-has-tabindex

## Rule Details

Enforces that any DOM element managing focus via `aria-activedescendant` is itself reachable by keyboard. An element using `aria-activedescendant` to indicate the currently focused descendant must either be a natively focusable element (`<input>`, `<button>`, `<select>`, etc.) or carry a `tabIndex` of `-1` or greater. A `tabIndex` of `-2` or lower removes the element from the focus tree, which defeats the activedescendant pattern.

The rule fires on a JSX opening element when **all** of the following hold:

- The element has an `aria-activedescendant` prop (case-insensitively; the
  boolean form `<div aria-activedescendant />` and explicit-undefined value
  `<div aria-activedescendant={undefined} />` both count as present).
- The resolved element name is in the HTML DOM set (custom React components,
  SVG-namespaced tags like `<svg:path>`, and dotted-namespace tags like
  `<Foo.Bar>` are skipped — the rule does not know what low-level element
  they render to).
- The element is not both inherently interactive (`<input>`, `<button>`,
  `<select>`, …) **and** missing a `tabIndex` prop. Inherently interactive
  elements rely on the browser-native tab order, so they don't need an
  explicit `tabIndex`.
- The element's resolved `tabIndex` value is less than `-1` (or cannot be
  statically resolved to a number `>= -1`).

Examples of **incorrect** code for this rule:

```jsx
<div aria-activedescendant={someID} />
<div aria-activedescendant={someID} tabIndex={-2} />
<ul aria-activedescendant={focusedId}><li>x</li></ul>
<section aria-activedescendant={x} tabIndex={-100}>content</section>
```

Examples of **correct** code for this rule:

```jsx
<div aria-activedescendant={someID} tabIndex={0} />
<div aria-activedescendant={someID} tabIndex={-1} />
<div aria-activedescendant={someID} tabIndex="0" />
<input aria-activedescendant={someID} />
<button aria-activedescendant={someID} />
<a href="#" aria-activedescendant={someID} />
<CustomComponent aria-activedescendant={someID} />
```

## Rule Options

This rule takes no options.

## Resources

- [WAI-ARIA — `aria-activedescendant`](https://www.w3.org/TR/wai-aria-1.2/#aria-activedescendant)
- [WAI-ARIA Authoring Practices — Managing Focus in Composites Using aria-activedescendant](https://www.w3.org/WAI/ARIA/apg/practices/keyboard-interface/#managingfocusincompositesusingaria-activedescendant)
- [MDN — `aria-activedescendant`](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Attributes/aria-activedescendant)
- [MDN — `tabindex`](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Global_attributes/tabindex)

## Original Documentation

- [eslint-plugin-jsx-a11y/aria-activedescendant-has-tabindex](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/aria-activedescendant-has-tabindex.md)
