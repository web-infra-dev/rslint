# scope

## Rule Details

The `scope` HTML attribute is only valid on `<th>` elements. Using it on any
other element is a no-op for assistive technology and signals a structural
mistake. This rule enforces that `scope` is used only on `<th>` elements,
matching axe-core's [`scope-attr-valid`](https://dequeuniversity.com/rules/axe/3.5/scope-attr-valid)
check (WCAG 1.3.1, 4.1.1).

The check is asymmetric on case:

- The `scope` attribute name is matched **case-insensitively** — `scope`,
  `SCOPE`, `Scope` all trigger.
- The element name is looked up in `aria-query`'s DOM map, which is keyed by
  the **lowercase** HTML element name. `<TH scope />` (uppercase tag) is
  silently skipped because `"TH"` is not a key in the map.

Examples of **incorrect** code for this rule:

```jsx
<div scope />
```

Examples of **correct** code for this rule:

```jsx
<th scope="col" />
<th scope={scope} />
```

## Settings

The rule reads `settings['jsx-a11y']` to resolve the effective element type
for a JSX tag. Resolution is two-step: the polymorphic prop runs first, then
the components map looks up the (possibly already-replaced) name.

- `polymorphicPropName` — name of a polymorphic prop (e.g. `"as"`) that
  remaps the element type. With `polymorphicPropName: "as"`,
  `<Box as="th" scope="col" />` resolves to `<th />` and is **not** flagged;
  `<Box as="div" scope />` resolves to `<div />` and is flagged.
- `polymorphicAllowList` — `string[]` restricting which raw component names
  may be remapped via the polymorphic prop. When omitted, every component
  may be remapped.
- `components` — a `{ ComponentName: "html-element" }` map. With
  `{ "TableHeader": "th" }`, `<TableHeader scope="row" />` is treated as
  `<th />` and is **not** flagged. The map can also remap a custom
  component to a non-`th` DOM element — e.g. `{ "Foo": "div" }` causes
  `<Foo scope />` to be reported as if it were `<div scope />`. When
  combined with `polymorphicPropName`, the lookup runs against the
  post-polymorphic name.

These mirror the upstream `eslint-plugin-jsx-a11y` settings exactly.

## Resources

- [WCAG 1.3.1 — Info and Relationships](https://www.w3.org/WAI/WCAG21/Understanding/info-and-relationships)
- [WCAG 4.1.1 — Parsing](https://www.w3.org/WAI/WCAG21/Understanding/parsing)
- [Deque University — `scope-attr-valid`](https://dequeuniversity.com/rules/axe/3.5/scope-attr-valid)
- [MDN — `<th>` `scope` attribute](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/th#scope)

## Original Documentation

- [eslint-plugin-jsx-a11y/scope](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/scope.md)
