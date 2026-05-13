# no-redundant-roles

Some HTML elements have native semantics that are implemented by the browser. This includes default/implicit ARIA roles. Setting an ARIA role that matches its default/implicit role is redundant since it is already set by the browser.

## Rule Details

The rule fires on a JSX opening element when **all** of the following hold:

- The element name resolves (through `settings['jsx-a11y'].polymorphicPropName` and `settings['jsx-a11y'].components`) to an HTML element that has an implicit ARIA role (`<button>` → `button`, `<nav>` → `navigation`, `<img>` → `img`, etc.).
- The element carries an explicit `role` attribute whose statically-literal value (lower-cased) equals that implicit role.
- The element is not in the allow-list for that tag (see [Rule Options](#rule-options)).

Non-literal role values (Identifiers, CallExpressions, template literals with substitutions, etc.) cannot be statically determined and never trigger the rule. The `role` attribute name is matched case-insensitively; the value is compared after `.toLowerCase()` so `role="BUTTON"` is the same as `role="button"`.

Some HTML elements have a context-sensitive implicit role:

- `<a>` / `<area>` / `<link>` — only acquire the `link` role when an `href` attribute is present.
- `<img>` — has implicit `img` UNLESS `alt=""` (decorative) or the literal `src` value contains `.svg`.
- `<input>` — the implicit role depends on `type`: `button`/`submit`/`reset`/`image` → `button`, `checkbox` → `checkbox`, `radio` → `radio`, `range` → `slider`, anything else (or absent) → `textbox`.
- `<select>` — `combobox` by default; `listbox` when `multiple` is truthy or `size > 1`.
- `<menu>` / `<menuitem>` — depend on the literal `type` value.

Examples of **incorrect** code for this rule:

```jsx
<button role="button" />
<img role="img" src="foo.jpg" />
<body role="document" />
<nav role="navigation" />
```

Examples of **correct** code for this rule:

```jsx
<div />
<button role="presentation" />
<MyComponent role="main" />
<img alt="" role="img" />
<img src="logo.svg" role="img" />
```

## Rule Options

The default options allow `nav` to carry an implicit `navigation` role, as is [advised by W3C](https://www.w3.org/WAI/GL/wiki/Using_HTML5_nav_element#Example:The_.3Cnav.3E_element). Options are provided as an object keyed by HTML element name; the value is an array of implicit ARIA roles that are allowed on the specified element.

```json
{ "jsx-a11y/no-redundant-roles": ["error", { "nav": ["navigation"] }] }
```

Specifying an entry for a key REPLACES the default for that key. To disable the built-in `nav` exception, set `nav` to an empty array:

```json
{ "jsx-a11y/no-redundant-roles": ["error", { "nav": [] }] }
```

Examples of **correct** code for this rule with `{ "ul": ["list"], "ol": ["list"] }`:

```json
{ "jsx-a11y/no-redundant-roles": ["error", { "ul": ["list"], "ol": ["list"] }] }
```

```jsx
<ul role="list" />
<ol role="list" />
```

## Resources

- [ARIA Spec — ARIA Adds Nothing to Default Semantics of Most HTML Elements](https://www.w3.org/TR/using-aria/#aria-does-nothing)
- [MDN — Identifying SVG as an image](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/img#identifying_svg_as_an_image)

## Original Documentation

- [eslint-plugin-jsx-a11y/no-redundant-roles](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/HEAD/docs/rules/no-redundant-roles.md)
